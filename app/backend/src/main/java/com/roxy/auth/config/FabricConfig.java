package com.roxy.auth.config;

import java.nio.file.Paths;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import com.roxy.auth.util.FabricUtils;

import io.grpc.ChannelCredentials;
import io.grpc.Grpc;
import io.grpc.ManagedChannel;
import io.grpc.TlsChannelCredentials;

@Configuration
public class FabricConfig {

  private final String mspId = "BusinessMSP";
  private final String channelName = "cross-domain-channel";
  private final String chaincodeName = "authcc";
  private final String networkPath = "/home/roxy3/blockchained_isolation_auth/fabric/test-network/organizations/peerOrganizations/business.com";

  @Bean
  public Gateway gateway() throws Exception {
    X509Certificate cert = FabricUtils
        .getCertificate(networkPath + "/users/Admin@business.com/msp/signcerts/Admin@business.com-cert.pem");
    PrivateKey key = FabricUtils.getPrivateKey(networkPath + "/users/Admin@business.com/msp/keystore/");

    return Gateway.newInstance()
        .identity(new X509Identity(mspId, cert))
        .signer(Signers.newPrivateKeySigner(key))
        .connection(grpcChannel())
        .connect();
  }

  @Bean
  public Contract contract(Gateway gateway) {
    return gateway.getNetwork(channelName).getContract(chaincodeName);
  }

  @Bean
  public ManagedChannel grpcChannel() throws Exception {
    ChannelCredentials creds = TlsChannelCredentials.newBuilder()
        .trustManager(Paths.get(networkPath + "/peers/peer0.business.com/tls/ca.crt").toFile())
        .build();
    return Grpc.newChannelBuilder("localhost:7051", creds)
        .overrideAuthority("peer0.business.com")
        .build();
  }
}
