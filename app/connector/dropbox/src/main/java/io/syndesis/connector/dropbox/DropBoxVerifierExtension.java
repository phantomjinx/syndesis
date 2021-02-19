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
package io.syndesis.connector.dropbox;

import java.util.Locale;
import java.util.Map;

import org.apache.camel.CamelContext;
import org.apache.camel.component.extension.verifier.DefaultComponentVerifierExtension;
import org.apache.camel.component.extension.verifier.ResultBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorHelper;

import com.dropbox.core.DbxException;
import com.dropbox.core.DbxRequestConfig;
import com.dropbox.core.v2.DbxClientV2;
import io.syndesis.connector.support.util.ConnectorOptions;

public class DropBoxVerifierExtension extends DefaultComponentVerifierExtension {

    public static final String ACCESS_TOKEN = "accessToken";
    public static final String CLIENT_IDENTIFIER = "clientIdentifier";

    protected DropBoxVerifierExtension(String defaultScheme, CamelContext context) {
        super(defaultScheme, context);
    }

    // *********************************
    // Parameters validation
    //
    @Override
    protected Result verifyParameters(Map<String, Object> parameters) {
        ResultBuilder builder = ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.PARAMETERS)
                .error(ResultErrorHelper.requiresOption(ACCESS_TOKEN, parameters))
                .error(ResultErrorHelper.requiresOption(CLIENT_IDENTIFIER, parameters));

        return builder.build();
    }

    // *********************************
    // Connectivity validation
    // *********************************
    @Override
    protected Result verifyConnectivity(Map<String, Object> parameters) {
        return ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.CONNECTIVITY)
                .error(parameters, DropBoxVerifierExtension::verifyCredentials).build();
    }

    private static void verifyCredentials(ResultBuilder builder, Map<String, Object> parameters) {

        String token = ConnectorOptions.extractOption(parameters, ACCESS_TOKEN);
        String clientId = ConnectorOptions.extractOption(parameters, CLIENT_IDENTIFIER);

        try {
            // Create Dropbox client
            DbxRequestConfig config = new DbxRequestConfig(clientId, Locale.getDefault().toString());
            DbxClientV2 client = new DbxClientV2(config, token);
            client.users().getCurrentAccount();
        } catch (DbxException e) {
            builder.error(ResultErrorBuilder
                    .withCodeAndDescription(VerificationError.StandardCode.AUTHENTICATION,
                            "Invalid client identifier and/or access token")
                    .parameterKey(ACCESS_TOKEN).parameterKey(CLIENT_IDENTIFIER).build());
        }

    }

}
