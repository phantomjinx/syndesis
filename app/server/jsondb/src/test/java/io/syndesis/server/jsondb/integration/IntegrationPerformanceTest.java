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
package io.syndesis.server.jsondb.integration;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Date;
import java.util.List;
import javax.sql.DataSource;
import org.apache.commons.lang3.time.StopWatch;
import org.h2.jdbcx.JdbcConnectionPool;
import org.junit.Before;
import org.junit.Test;
import org.skife.jdbi.v2.DBI;
import org.springframework.core.io.DefaultResourceLoader;
import io.syndesis.common.model.DataShape;
import io.syndesis.common.model.DataShapeKinds;
import io.syndesis.common.model.Dependency;
import io.syndesis.common.model.action.ConnectorAction;
import io.syndesis.common.model.action.ConnectorDescriptor;
import io.syndesis.common.model.connection.Connection;
import io.syndesis.common.model.connection.Connector;
import io.syndesis.common.model.integration.Flow;
import io.syndesis.common.model.integration.Integration;
import io.syndesis.common.model.integration.Step;
import io.syndesis.common.model.integration.StepKind;
import io.syndesis.common.util.EventBus;
import io.syndesis.common.util.KeyGenerator;
import io.syndesis.common.util.cache.Cache;
import io.syndesis.common.util.cache.CacheManager;
import io.syndesis.server.dao.manager.DataAccessObject;
import io.syndesis.server.dao.manager.DataManager;
import io.syndesis.server.dao.manager.EncryptionComponent;
import io.syndesis.server.jsondb.dao.ConnectionJsonDbDao;
import io.syndesis.server.jsondb.dao.ConnectorJsonDbDao;
import io.syndesis.server.jsondb.dao.IntegrationJsonDbDao;
import io.syndesis.server.jsondb.impl.SqlJsonDB;

public class IntegrationPerformanceTest {

    private StopWatch sw = new StopWatch();
    private DataManager dataManager;
    private SqlJsonDB jsonDB;

    private Connector odataConnector;
    private Connection odataConnection;

    private void startTimer() {
        sw.reset();
        sw.start();
    }

    private long stopTimer() {
        sw.stop();
        return sw.getTime();
    }

    private Connector createODataConnector() {
        Connector.Builder builder = new Connector.Builder()
            .id("odata")
            .name("OData")
            .componentScheme("olingo4")
            .description("Communicate with an OData service")
            .addDependency(Dependency.maven("org.apache.camel:camel-olingo4:latest"))
            .putConfiguredProperty("service-uri", "http://www.anything.org/odata.svc");

        return builder.build();
    }

    private Connection createODataConnection() {
        return new Connection.Builder()
            .putConfiguredProperty("basicPassword", "password")
            .putConfiguredProperty("basicUserName", "user")
            .connectorId(odataConnector.getId().get())
            .id(KeyGenerator.createKey())
            .name("MyODataConn")
            .connector(odataConnector)
            .build();
    }

    private Step createODataStep() {
        return new Step.Builder()
        .stepKind(StepKind.endpoint)
        .action(new ConnectorAction.Builder()
                .description("Read resource entities from the server subject to keyPredicates")
                .id("io.syndesis:odata-read-connector-from")
                .name("Read")
                .descriptor(new ConnectorDescriptor.Builder()
                           .componentScheme("olingo4")
                           .putConfiguredProperty("method", "read")
                           .putConfiguredProperty("connection-direction", "to")
                           .addConnectorCustomizer("io.syndesis.connector.odata.customizer.ODataReadToCustomizer")
                           .connectorFactory("io.syndesis.connector.odata.component.ODataComponentFactory")
                           .inputDataShape(new DataShape.Builder()
                                           .kind(DataShapeKinds.JSON_SCHEMA)
                                           .build())
                           .outputDataShape(new DataShape.Builder()
                                            .kind(DataShapeKinds.JSON_INSTANCE)
                                            .build())
                           .build())
               .build())
        .connection(
                        odataConnection)
        .putConfiguredProperty("resource-path", "/myResource")
        .build();
    }

    private Step createLogStep() {
        return new Step.Builder()
            .stepKind(StepKind.endpoint)
            .action(new ConnectorAction.Builder()
                    .descriptor(new ConnectorDescriptor.Builder()
                                .componentScheme("log")
                                .build())
                    .build())
            .build();
    }

    private Integration createIntegration(String id, String name, Step... steps) {

        Flow.Builder flowBuilder = new Flow.Builder();
        for (Step step : steps) {
            flowBuilder.addStep(step);
        }

        return new Integration.Builder()
            .id(id)
            .name(name)
            .addTags("log", "odata")
            .addFlow(flowBuilder.build())
            .build();
    }

    @Before
    public void setup() {
        odataConnector = createODataConnector();
        odataConnection = createODataConnection();

        EventBus bus = mock(EventBus.class);
        final DataSource dataSource = JdbcConnectionPool.create("jdbc:h2:mem:", "sa", "password");
        final DBI dbi = new DBI(dataSource);
        jsonDB = new SqlJsonDB(dbi, bus);
        jsonDB.createTables();

        @SuppressWarnings( "unchecked" )
        Cache<Object, Object> cache = mock(Cache.class);
        CacheManager cacheMgr = mock(CacheManager.class);
        when(cacheMgr.getCache(any(String.class), any(boolean.class))).thenReturn(cache);

        IntegrationJsonDbDao integrationDAO = new IntegrationJsonDbDao(jsonDB);
        ConnectorJsonDbDao connectorDAO = new ConnectorJsonDbDao(jsonDB);
        ConnectionJsonDbDao connectionDAO = new ConnectionJsonDbDao(jsonDB);
        DataAccessObject<?>[] daos = { integrationDAO, connectionDAO, connectorDAO };

        //Create Data Manager
        dataManager = new DataManager(cacheMgr,
                                      Arrays.asList(daos),
                                      null, new EncryptionComponent(null), new DefaultResourceLoader(), null);
        dataManager.init();
        dataManager.resetDeploymentData();
    }

    @Test
    public void testCreateAndReadIntegrations() throws Exception {
        int testCount = 10;
        String[] ids = new String[testCount];
        String[] names = new String[testCount];

        //
        // Create some integrations
        //
        List<Integration> createdInts = new ArrayList<>();
        for (int i = 0; i < testCount; ++i) {
            ids[i] = KeyGenerator.createKey() + "-" + i;
            names[i] = "MyODataInt" + "-" + i;
            createdInts.add(dataManager.create(createIntegration(ids[i], names[i], createODataStep(), createLogStep())));
        }
        assertEquals(testCount, createdInts.size());

        //
        // Confirm integrations created by reading
        //
        startTimer();
        List<Integration> fetchedInts = dataManager.fetchAll(Integration.class).getItems();
        long timeTaken = stopTimer();
        assertEquals(createdInts.size(), fetchedInts.size());

        List<Integration> removeInts = new ArrayList<>(fetchedInts);
        for (Integration c : createdInts) {
            if (fetchedInts.contains(c)) {
                removeInts.remove(c);
            }
        }

        //
        // TODO
        // Failing since the target url is still resident in the flow/step/connection field
        //

        assertTrue(removeInts.isEmpty());

//        for (Integration integration : integrations) {
//            System.out.println(integration.getName() + ": " + integration.getKind());
//        }
//        System.out.println("FetchAll Time Taken: " + sw.getTime());
//
//        sw.reset();
//
        sw.start();
        Integration integration = dataManager.fetch(Integration.class, ids[0]);
        sw.stop();
        assertNotNull(integration);
        assertEquals(names[0], integration.getName());

        System.out.println("Fetch Time Taken: " + sw.getTime());

        Connection connection = dataManager.fetch(Connection.class, "5");
        assertNotNull(connection);
        System.out.println("=== Found connection! ===");

        List<Connection> connections = dataManager.fetchAll(Connection.class).getItems();
        assertTrue(connections.size() > 0);
        for (Connection c : connections) {
            assertTrue(c.getId().isPresent());
            assertTrue(c.getName() != null);
            System.out.println(c.getId().orElse("NULL") + "\t" + c.getName() + ": " + c.getKind());
        }

        String path0 = "/integrations/:" + ids[0] + "/";
        List<String> paths = jsonDB.list(path0);
        assertEquals(1, paths.size());
        assertEquals(path0, paths.get(0));

        int count = jsonDB.count(path0);
        assertEquals(1, count);
    }
}
