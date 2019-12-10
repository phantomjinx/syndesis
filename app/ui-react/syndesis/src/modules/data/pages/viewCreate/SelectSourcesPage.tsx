import { SchemaNodeInfo, Virtualization } from '@syndesis/models';
import { CreateViewHeader, ViewCreateLayout } from '@syndesis/ui';
import { useRouteData } from '@syndesis/utils';
import * as React from 'react';
import resolvers from '../../../resolvers';
import { ConnectionSchemaContent } from '../../shared';
import { ConnectionPreviewSchema } from '../../shared/ConnectionPreviewSchema';

/**
 * @param virtualizationId - the ID of the virtualization for the wizard
 */
export interface ISelectSourcesRouteParams {
  virtualizationId: string;
}

/**
 * @param virtualization - the virtualization for the wizard.
 */
export interface ISelectSourcesRouteState {
  virtualization: Virtualization;
}

export interface ISelectSourcesPageProps {
  handleNodeSelected: (
    connectionName: string,
    name: string,
    teiidName: string,
    nodePath: string[]
  ) => void;
  handleNodeDeselected: (connectionName: string, teiidName: string) => void;
  selectedSchemaNodes: SchemaNodeInfo[];
}

export const SelectSourcesPage: React.FunctionComponent<
  ISelectSourcesPageProps
> = props => {
  const { state } = useRouteData<null, ISelectSourcesRouteState>();

  const schemaNodeInfo: SchemaNodeInfo[] = props.selectedSchemaNodes;
  const virtualization = state.virtualization;

  return (
    <ViewCreateLayout
      header={<CreateViewHeader 
        step={1}
        cancelHref={resolvers.data.virtualizations.views.root({
          virtualization,
        })}
        nextHref={resolvers.data.virtualizations.views.createView.selectName({
          schemaNodeInfo,
          virtualization,
        })}
        isNextDisabled={props.selectedSchemaNodes.length > 1}
        isNextLoading={false}
        isLastStep={false}
        />}
      content={
        <ConnectionSchemaContent
          onNodeSelected={props.handleNodeSelected}
          onNodeDeselected={props.handleNodeDeselected}
          selectedSchemaNodes={props.selectedSchemaNodes}
        />
      }
      preview={<ConnectionPreviewSchema
        selectedSchemaNodes={props.selectedSchemaNodes}
      />}
    />
  );
};
