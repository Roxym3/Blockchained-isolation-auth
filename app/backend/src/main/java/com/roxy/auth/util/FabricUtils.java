package com.roxy.auth.util;

import java.io.BufferedReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.util.stream.Stream;

import org.hyperledger.fabric.client.identity.Identities;

public class FabricUtils {

    /**
     * 读取指定路径的 X.509 证书文件
     */
    public static X509Certificate getCertificate(String certPath) throws Exception {
        try (BufferedReader reader = Files.newBufferedReader(Paths.get(certPath))) {
            return Identities.readX509Certificate(reader);
        }
    }

    /**
     * 自动从目录中读取第一个私钥文件 (针对 cryptogen 生成的随机文件名)
     */
    public static PrivateKey getPrivateKey(String keyPathDir) throws Exception {
        try (Stream<Path> keyFiles = Files.list(Paths.get(keyPathDir))) {
            Path keyPath = keyFiles.findFirst().orElseThrow(() -> 
                new RuntimeException("在目录中未找到私钥文件: " + keyPathDir));
            try (BufferedReader reader = Files.newBufferedReader(keyPath)) {
                return Identities.readPrivateKey(reader);
            }
        }
    }
}