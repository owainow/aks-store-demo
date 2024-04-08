#!/bin/bash

##########################################################
# Create the custom-values.yaml file with base images
##########################################################

cat << EOF > custom-values.yaml
namespace: ${AZURE_AKS_NAMESPACE}
productService:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/product-service
storeAdmin:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/store-admin
storeFront:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/store-front
virtualCustomer:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/virtual-customer
virtualWorker:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/virtual-worker
EOF

###########################################################
# Add ai-servie if Azure OpenAI endpoint is provided
###########################################################

if [ -n "${AZURE_OPENAI_ENDPOINT}" ]; then
  cat << EOF >> custom-values.yaml
aiService:
  image:
      repository: ${AZURE_REGISTRY_URI}/aks-store-demo/ai-service
  create: true
  modelDeploymentName: ${AZURE_OPENAI_MODEL_NAME}
  openAiEndpoint: ${AZURE_OPENAI_ENDPOINT}
  useAzureOpenAi: true
EOF

  # If Azure identity exists, use it, otherwise use the Azure OpenAI API key
  if [ -n "${AZURE_IDENTITY_CLIENT_ID}" ]; then
    cat << EOF >> custom-values.yaml
  useAzureAd: true
  managedIdentityClientId: ${AZURE_IDENTITY_CLIENT_ID}
EOF
  else
    cat << EOF >> custom-values.yaml
  openAiKey: $(az keyvault secret show --name ${AZURE_OPENAI_KEY} --vault-name ${AZURE_KEY_VAULT_NAME} --query value -o tsv)
EOF
  fi
fi

###########################################################
# Add order-service
###########################################################

cat << EOF >> custom-values.yaml
orderService:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/order-service
EOF

# Add Azure Service Bus to order-service if provided
if [ -n "${AZURE_SERVICE_BUS_HOST}" ]; then
  cat << EOF >> custom-values.yaml
  queueHost: ${AZURE_SERVICE_BUS_HOST}
  queuePort: "5671"
  queueTransport: "tls"
  queueUsername: ${AZURE_SERVICE_BUS_SENDER_NAME}
  queuePassword: $(az keyvault secret show --name ${AZURE_SERVICE_BUS_SENDER_KEY} --vault-name ${AZURE_KEY_VAULT_NAME} --query value -o tsv)
EOF
fi

###########################################################
# Add makeline-service
###########################################################

cat << EOF >> custom-values.yaml
makelineService:
  image:
    repository: ${AZURE_REGISTRY_URI}/aks-store-demo/makeline-service
EOF

# Add Azure Service Bus to makeline-service if provided
if [ -n "${AZURE_SERVICE_BUS_URI}" ]; then
  cat << EOF >> custom-values.yaml
  orderQueueUri: ${AZURE_SERVICE_BUS_URI}
  orderQueueUsername: ${AZURE_SERVICE_BUS_LISTENER_NAME}
  orderQueuePassword: $(az keyvault secret show --name ${AZURE_SERVICE_BUS_LISTENER_KEY} --vault-name ${AZURE_KEY_VAULT_NAME} --query value -o tsv)
EOF
fi

# Add Azure Cosmos DB to makeline-service if provided
if [ -n "${AZURE_COSMOS_DATABASE_URI}" ]; then
  cat << EOF >> custom-values.yaml
  orderDBApi: ${AZURE_DATABASE_API}
  orderDBUri: ${AZURE_COSMOS_DATABASE_URI}
  orderDBUsername: ${AZURE_COSMOS_DATABASE_NAME}
  orderDBPassword: $(az keyvault secret show --name ${AZURE_COSMOS_DATABASE_KEY} --vault-name ${AZURE_KEY_VAULT_NAME} --query value -o tsv)
EOF
fi

###########################################################
# Do not deploy RabbitMQ when using Azure Service Bus
###########################################################
if [ -n "${AZURE_SERVICE_BUS_HOST}" ]; then
  cat << EOF >> custom-values.yaml
useRabbitMQ: false
EOF
fi

###########################################################
# Do not deploy MongoDB when using Azure Cosmos DB
###########################################################
if [ -n "${AZURE_COSMOS_DATABASE_URI}" ]; then
  cat << EOF >> custom-values.yaml
useMongoDB: false
EOF
fi