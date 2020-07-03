package io.syndesis.server.controller.integration.camelk.customizer;

import io.syndesis.server.controller.integration.camelk.crd.Integration;
import io.syndesis.server.controller.integration.camelk.crd.TraitSpec;
import io.syndesis.server.openshift.Exposure;
import io.syndesis.server.openshift.OpenShiftServiceImpl;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.EnumSet;
import java.util.Map;

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

}
