/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package io.syndesis.connector.odata.verifier;

import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.PublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;
import org.junit.Test;
import org.keycloak.RSATokenVerifier;
import org.keycloak.common.VerificationException;
import org.keycloak.representations.AccessToken;

public class TestBlah {

    @Test
    public void testblah() throws VerificationException {
        System.out.println("Hello");
        String tokenString = "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJXTHRhald5ZmRzVjJYREF3cjhKRWxMaW5DTThKS1Y3bndiUEtwNGoxNkZrIn0.eyJleHAiOjE1OTI1MTYzMDcsImlhdCI6MTU5MjUxNjAwNywianRpIjoiYWRlY2NkYTgtZWUxYS00NjU4LWIyMWEtNTA1NTNlYWY1MjMxIiwiaXNzIjoiaHR0cHM6Ly9rZXljbG9hay5kYXNoL2F1dGgvcmVhbG1zL3N5bmRlc2lzIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6ImYxNmQ5MzllLWU5ZmUtNDYyYi05MTAwLTMxNGFlMGY2Njg4MiIsInR5cCI6IkJlYXJlciIsImF6cCI6InN5bmRlc2lzIiwic2Vzc2lvbl9zdGF0ZSI6ImQ4YjgyMjk3LTZiNGUtNDQ4MS1hNTgwLWMyYjI2NjMwMDNhZCIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOlsiaHR0cHM6Ly9zeW5kZXNpcy5kYXNoIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlcnMiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6InByb2ZpbGUgZW1haWwiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiVXNlcm5hbWUiOiJwaGFudG9tamlueCIsIm5hbWUiOiJQYXVsIFJpY2hhcmRzb24iLCJncm91cHMiOlsiL3N5bnVzZXJzIl0sInByZWZlcnJlZF91c2VybmFtZSI6InBoYW50b21qaW54IiwiZ2l2ZW5fbmFtZSI6IlBhdWwiLCJmYW1pbHlfbmFtZSI6IlJpY2hhcmRzb24iLCJlbWFpbCI6InAuZy5yaWNoYXJkc29uQHBoYW50b21qaW54LmNvLnVrIn0.3cbvc34VfCxjtgktpu_1TwACvD85Ya3qCra_SVvaWV55xIeacAeg_5NVb-XL-nQPIyu-YA3gpWj7uTlY-tXI_doqKRCM_7CcvjBHzFB2t6wpdc3amxgOje1QBHhdQHVSiA8Rila5ztl0s2MdZGNfnXU8xSaIHn7KwO4KsVfvA4a3Q29d9vtwcC4zqvcRurbkQF4qIGNbo0fXFRnWS2gKlRCh1CT2sIFMUP4nUrFPU3_sTZ0IXiITykteXzYs5ilkasaGTTMvru-ERhElqhQv9U0SfFBG3JyNUL9IKLHeYcfVeMNW6t7DuoTlz6zhQy4E67Q4ZoKIGW_xOVyDGHcwtw";
        String publicKeyString = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqjklywp9AfPJBI3KbfD0N9zpitQnOq2utL/CT9W+XQyb9L740o18dZf1b8K4DR6nBoBhzRPu9BtBJp4sRY6uY5z91Wk/hMOhj8Y577Q2Tk5pFyTh+plxstpxqUKWvn9nipROcVX/BVQjlyuGz4p8AyoSrSSDLlWX1Hc8qH4kGdXPDEQf05dJ8FkUSFkcYmjCKXZgtMe4k4rgdKbk3e5L6LGfpJ/AWzNLoAYG5p+FY46v+45csmc0Xqg9qXnU++u4fWzXjHIQhhyhQxHrTjdTpyEn9lAPvRy8spPrjEokdgO8yIYufhS907dMwe8BLfbriosXi429rPb1RXSz5n8OzQIDAQAB";

        RSATokenVerifier verifier = RSATokenVerifier.create(tokenString);
        PublicKey publicKey = toPublicKey(publicKeyString);
        
        System.out.println("Algorithm: " + publicKey.getAlgorithm());
        System.out.println("Public Key: " + publicKey.getFormat());

        AccessToken token = verifier.realmUrl("https://keycloak.dash/auth/realms/syndesis") //
          .publicKey(publicKey) //
          .getToken();
        
        System.out.println("email: " + token.getEmail());
        System.out.println("family name: " + token.getFamilyName());
        System.out.println("id: " + token.getId());
        System.out.println("name: " + token.getName());
        System.out.println("address: " + token.getAddress());
        System.out.println("Origins: " + token.getAllowedOrigins());
        System.out.println("AuthTime: " + token.getAuthTime());
        System.out.println("birthdata: " + token.getBirthdate());
        System.out.println("issued for: " + token.getIssuedFor());
        System.out.println("scope: " + token.getScope());
        System.out.println("Realm Access: " + token.getRealmAccess());
    }
    
    public PublicKey toPublicKey(String publicKeyString) {
        try {
          byte[] publicBytes = Base64.getDecoder().decode(publicKeyString);
          X509EncodedKeySpec keySpec = new X509EncodedKeySpec(publicBytes);
          KeyFactory keyFactory = KeyFactory.getInstance("RSA");
          return keyFactory.generatePublic(keySpec);
        } catch (NoSuchAlgorithmException | InvalidKeySpecException e) {
          return null;
        }
      }
}
