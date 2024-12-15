FROM confluentinc/cp-kafka-connect:latest.arm64

# Download and install the JDBC Source Connector
RUN confluent-hub install --no-prompt confluentinc/kafka-connect-jdbc:latest

# Set the plugin path
ENV CONNECT_PLUGIN_PATH=/usr/share/java,/usr/share/confluent-hub-components
