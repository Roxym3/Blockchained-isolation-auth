package com.roxy.auth.controller;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.EndorseException;
import org.hyperledger.fabric.client.SubmitException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/auth")
public class AuthController {

  @Autowired
  private Contract contract;

  @PostMapping("/issue")
  public ResponseEntity<String> issueTicket(
      @RequestParam String ticketId,
      @RequestParam String targetDomain,
      @RequestParam String owner,
      @RequestParam String expiryTime) {
    try {
      // 🌟 1.x 的标准调用：直接 submit，由网关节点去搞定背书策略
      contract.submitTransaction("IssueTicket", ticketId, targetDomain, owner, expiryTime);

      return ResponseEntity.ok("票据签发成功！TicketID: " + ticketId);

    } catch (EndorseException e) {
      // 如果报 no combination of peers，说明是策略没过
      return ResponseEntity.status(500).body(" 背书阶段失败: " + e.getMessage());
    } catch (SubmitException e) {
      return ResponseEntity.status(500).body(" 提交阶段失败: " + e.getMessage());
    } catch (Exception e) {
      return ResponseEntity.status(500).body(" 内部系统错误: " + e.getMessage());
    }
  }
}
