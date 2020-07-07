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
package io.syndesis.server.controller.integration.camelk.customizer;

import java.util.Collections;
import java.util.EnumSet;
import java.util.Map;
import java.util.Optional;

import io.syndesis.common.model.integration.IntegrationDeployment;
import io.syndesis.server.controller.integration.camelk.crd.ConfigurationSpec;
import io.syndesis.server.controller.integration.camelk.crd.Integration;
import io.syndesis.server.controller.integration.camelk.crd.TraitSpec;
import io.syndesis.server.openshift.Exposure;
import io.syndesis.server.openshift.OpenShiftServiceImpl;
import org.springframework.stereotype.Component;

@Component
public class JacksonCustomizer extends AbstractTraitCustomizer {

    @Override
    protected Map<String, TraitSpec> computeTraits(Integration integration, EnumSet<Exposure> exposure) {
        return Collections.singletonMap(
            "jvm",
            new TraitSpec.Builder()
                .putConfiguration("options", OpenShiftServiceImpl.JACKSON_OPTIONS)
                .build()
        );
    }

    @Override
    public Integration customize(IntegrationDeployment deployment, Integration integration, EnumSet<Exposure> exposure) {
        Optional<ConfigurationSpec> javaOptions = integration.getSpec().getConfiguration().stream()
            .filter(configurationSpec ->
                "env".equals(configurationSpec.getType()) &&
                    configurationSpec.getValue() != null &&
                    configurationSpec.getValue().contains("JAVA_OPTIONS"))
            .findFirst();
        if (javaOptions.isPresent()) {
            ConfigurationSpec cs = javaOptions.get();
            String options = cs.getValue() + " " + OpenShiftServiceImpl.JACKSON_OPTIONS;
            integration.getSpec().getConfiguration().remove(javaOptions.get());
            integration.getSpec().getConfiguration().add(ConfigurationSpec.of("env", options));
        } else {
            String options = "JAVA_OPTIONS=" + OpenShiftServiceImpl.JACKSON_OPTIONS;
            integration.getSpec().getConfiguration().add(ConfigurationSpec.of("env", options));
        }
        return integration;
    }
}
