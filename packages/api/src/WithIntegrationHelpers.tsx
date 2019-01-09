import {
  Action,
  ActionDescriptor,
  Connection,
  Integration,
  Step,
} from '@syndesis/models';
import { key } from '@syndesis/utils';
import produce from 'immer';
import * as React from 'react';
import { ApiContext, IApiContext } from './ApiContext';
import { callFetch } from './callFetch';
import {
  deserializeIntegration,
  serializeIntegration,
} from './integrationHelpers';

export const NEW_INTEGRATION = {
  name: '',
  tags: [],
} as Integration;
export const NEW_INTEGRATION_ID = 'new-integration';

type AddConnection = (
  integration: Integration,
  connection: Connection,
  action: Action,
  flow: number,
  position: number,
  configuredProperties: {
    [name: string]: string;
  }
) => Promise<Integration>;
type UpdateConnection = (
  integration: Integration,
  connection: Connection,
  action: Action,
  flow: number,
  position: number,
  configuredProperties: {
    [name: string]: string;
  }
) => Promise<Integration>;

export interface IWithIntegrationHelpersChildrenProps {
  addConnection: AddConnection;
  updateConnection: UpdateConnection;
  getEmptyIntegration(integration?: Integration): Integration;
  saveIntegration(integration: Integration): Promise<Integration>;
  setName(integration: Integration, name: string): Integration;
  createDraft(value: Integration): Promise<string>;
  getCreationDraft(): Integration;
  getDraft(id: string): Promise<Integration>;
  getStep(value: Integration, flow: number, step: number): Step;
  setCreationDraft(value: Integration): void;
  setDraft(id: string, value: Integration): Promise<void>;
}

export interface IWithIntegrationHelpersProps {
  children(props: IWithIntegrationHelpersChildrenProps): any;
}

export class WithIntegrationHelpersWrapped extends React.Component<
  IWithIntegrationHelpersProps & IApiContext
> {
  constructor(props: IWithIntegrationHelpersProps & IApiContext) {
    super(props);
    this.addConnection = this.addConnection.bind(this);
    this.getEmptyIntegration = this.getEmptyIntegration.bind(this);
    this.saveIntegration = this.saveIntegration.bind(this);
    this.setName = this.setName.bind(this);
    this.updateConnection = this.updateConnection.bind(this);
    this.makeLocalStorageId = this.makeLocalStorageId.bind(this);
    this.createDraft = this.createDraft.bind(this);
    this.getDraft = this.getDraft.bind(this);
    this.setDraft = this.setDraft.bind(this);
    this.getCreationDraft = this.getCreationDraft.bind(this);
    this.setCreationDraft = this.setCreationDraft.bind(this);
    this.getStep = this.getStep.bind(this);
  }

  public async getActionDescriptor(
    connectionId: string,
    actionId: string,
    configuredProperties: any
  ): Promise<ActionDescriptor> {
    const response = await callFetch({
      body: configuredProperties,
      headers: this.props.headers,
      method: 'POST',
      url: `${
        this.props.apiUri
      }/connections/${connectionId}/actions/${actionId}`,
    });
    if (!response.ok) {
      throw new Error(response.statusText);
    }
    return (await response.json()) as ActionDescriptor;
  }

  public async addConnection(
    integration: Integration,
    connection: Connection,
    action: Action,
    flow: number,
    position: number,
    configuredProperties: {
      [name: string]: string;
    }
  ): Promise<Integration> {
    const actionDescriptor = await this.getActionDescriptor(
      connection.id!,
      action.id!,
      configuredProperties
    );
    return produce(integration, draft => {
      if (!draft.flows) {
        draft.flows = [];
      }
      if (!draft.flows[flow]) {
        draft.flows[flow] = {
          id: key(),
          name: '',
          steps: [],
        };
      }
      if (!draft.flows[flow].steps) {
        draft.flows[flow].steps = [];
      }
      const step: Step = {
        action,
        configuredProperties,
        connection,
        id: draft.flows[flow].id,
      };
      step.action!.descriptor = actionDescriptor;
      step.stepKind = 'endpoint';
      draft.flows[flow].steps!.splice(position, 0, step);
      draft.tags = Array.from(new Set([...(draft.tags || []), connection.id!]));
    });
  }

  public async updateConnection(
    integration: Integration,
    connection: Connection,
    action: Action,
    flow: number,
    position: number,
    configuredProperties: {
      [name: string]: string;
    }
  ): Promise<Integration> {
    const actionDescriptor = await this.getActionDescriptor(
      connection.id!,
      action.id!,
      configuredProperties
    );
    return produce(integration, draft => {
      if (!draft.flows) {
        draft.flows = [];
      }
      if (!draft.flows[flow]) {
        draft.flows[flow] = {
          id: key(),
          name: '',
          steps: [],
        };
      }
      if (!draft.flows[flow].steps) {
        draft.flows[flow].steps = [];
      }
      const step: Step = {
        action,
        configuredProperties,
        connection,
        id: draft.flows[flow].id,
      };
      step.action!.descriptor = actionDescriptor;
      step.stepKind = 'endpoint';
      draft.flows[flow].steps![position] = step;
    });
  }

  public getEmptyIntegration(): Integration {
    return NEW_INTEGRATION;
  }

  public async saveIntegration(integration: Integration): Promise<Integration> {
    const response = await callFetch({
      body: integration,
      headers: this.props.headers,
      method: integration.id ? 'PUT' : 'POST',
      url: integration.id
        ? `${this.props.apiUri}/integrations/${integration.id}`
        : `${this.props.apiUri}/integrations`,
    });
    if (!response.ok) {
      throw new Error(response.statusText);
    }
    return (await response.json()) as Integration;
  }

  public setName(integration: Integration, name: string): Integration {
    return produce(integration, nextIntegration => {
      nextIntegration.name = name;
    });
  }

  public makeLocalStorageId(id: string) {
    return `iec-${id}`;
  }

  public async createDraft(value: Integration) {
    // TODO: this should be handled by the BE
    const id = value.id || NEW_INTEGRATION_ID;
    await this.setDraft(id, value);
    return Promise.resolve(id);
  }

  public getDraft(id: string): Promise<Integration> {
    // TODO: this should be handled by the BE
    const serializedIntegration = localStorage.getItem(
      this.makeLocalStorageId(id)
    );
    if (!serializedIntegration) {
      throw new Error(`There is no draft for id ${id}`);
    }
    return Promise.resolve(deserializeIntegration(serializedIntegration));
  }

  public setDraft(id: string, value: Integration): Promise<void> {
    // TODO: this should be handled by the BE
    localStorage.setItem(
      this.makeLocalStorageId(id),
      serializeIntegration(value)
    );
    return Promise.resolve();
  }

  public getCreationDraft(): Integration {
    const serializedIntegration = localStorage.getItem(
      this.makeLocalStorageId(NEW_INTEGRATION_ID)
    );
    if (!serializedIntegration) {
      throw new Error('There is no creation draft');
    }
    return deserializeIntegration(serializedIntegration);
  }

  public setCreationDraft(value: Integration): void {
    localStorage.setItem(
      this.makeLocalStorageId(NEW_INTEGRATION_ID),
      serializeIntegration(value)
    );
  }

  public getStep(value: Integration, flow: number, step: number): Step {
    try {
      return value.flows![flow].steps![step];
    } catch (e) {
      throw new Error(
        `Can't find a step in position flow:${flow} step:${step}`
      );
    }
  }

  public render() {
    return this.props.children({
      addConnection: this.addConnection,
      createDraft: this.createDraft,
      getCreationDraft: this.getCreationDraft,
      getDraft: this.getDraft,
      getEmptyIntegration: this.getEmptyIntegration,
      getStep: this.getStep,
      saveIntegration: this.saveIntegration,
      setCreationDraft: this.setCreationDraft,
      setDraft: this.setDraft,
      setName: this.setName,
      updateConnection: this.updateConnection,
    });
  }
}

export const WithIntegrationHelpers: React.FunctionComponent<
  IWithIntegrationHelpersProps
> = props => (
  <ApiContext.Consumer>
    {apiContext => <WithIntegrationHelpersWrapped {...props} {...apiContext} />}
  </ApiContext.Consumer>
);
