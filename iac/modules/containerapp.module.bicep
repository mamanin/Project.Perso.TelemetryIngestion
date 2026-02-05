// =======================================================================
// Container App Module
// -----------------------------------------------------------------------
// Module: containerapp.module.bicep
// Description: Deploys an Azure Container App
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.app/containerapps
// =======================================================================

import { KeyValueType } from '../types/keyvalue.type.bicep'
import { ContainerAppProbesType } from '../types/containerapp.probes.type.bicep'
import { ContainerAppScaleRule } from '../types/containerapp.rule.type.bicep'
import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Container App')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The user assigned identity to use for the Container App')
param managedIdentityId string

@description('The active revisions mode for the Container App')
param activeRevisionsMode 'Single' | 'Multiple' = 'Multiple'

@description('The Container App Environment ID')
param containerAppEnvironmentId string

@description('The Azure Container Registry server URL')
param containerServer string

@description('The container image to deploy')
param containerImage string

@description('The number of CPU cores the container can use')
@allowed([
  '0.25'
  '0.5'
  '0.75'
  '1.0'
  '1.25'
  '1.5'
  '1.75'
  '2.0'
])
param cpuCore string = '0.25'

@description('The amount of memory (in Gi) allocated to the container')
param memorySize string = '0.5'

@description('Scaling rules for the Container App')
param scaleRules ContainerAppScaleRule[] = []

@description('Min replicas')
param minReplicas int = 0

@description('Max replicas')
param maxReplicas int = 2

@description('Environment variables for the container')
param env KeyValueType[] = []

@description('Environment variables for the container')
param probes ContainerAppProbesType[] = []

@description('Target port for the application')
param applicationPort int

var containerAppName = BuildResourceName(prefix, 'aca', number)

resource containerApp 'Microsoft.App/containerApps@2025-10-02-preview' = {
  name: containerAppName
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${managedIdentityId}': {}
    }
  }
  tags: tags
  properties: {
    managedEnvironmentId: containerAppEnvironmentId
    configuration: {
      activeRevisionsMode: activeRevisionsMode
      ingress: {
        external: true
        targetPort: applicationPort
        allowInsecure: false
        clientCertificateMode: 'require'
      }
      registries: [
        {
          server: containerServer
          identity: managedIdentityId
        }
      ]
    }
    template: {
      containers: [
        {
          name: containerAppName
          image: '${containerServer}${containerImage}'
          resources: {
            cpu: json(cpuCore)
            memory: '${memorySize}Gi'
          }
          env: env
          probes: [for prob in probes: {
            type: prob.type
            initialDelaySeconds: prob.?initialDelaySeconds ?? 1
            periodSeconds: prob.periodSeconds
            timeoutSeconds: prob.?timeoutSeconds ?? 1
            failureThreshold: prob.?failureThreshold ?? 3
            tcpSocket: prob.scheme == 'TCP' ? {
              port: prob.?port ?? applicationPort
            } : null
            httpGet: prob.scheme == 'HTTP' ? {
              port: applicationPort
              path: prob.?path ?? '/'
              scheme: prob.scheme
            } : null
          }]
        }
      ]
      scale: {
        minReplicas: minReplicas
        maxReplicas: maxReplicas
        rules: scaleRules
      }
    }
  }
}

@description('The resource ID of the Container App')
output id string = containerApp.id

@description('The name of the Container App')
output name string = containerApp.name

@description('The FQDN of the Container App')
output fqdn string = containerApp.properties.configuration.ingress.fqdn
