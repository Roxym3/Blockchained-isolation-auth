package com.roxy.auth;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class BackendApplication {
    public static void main(String[] args) {
        // 启动 Spring Boot 应用
        SpringApplication.run(BackendApplication.class, args);
        System.out.println(" 跨域认证后端服务已启动！可以开始调用区块链网络了！");
    }
}