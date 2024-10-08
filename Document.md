## Overview
The Log Listener is designed to listen for log data from various sources, enrich the logs with GeoIP information, and store the processed logs in a Manticore search engine. The application supports both UDP and TCP protocols for receiving log data and includes a mechanism for handling configuration, processing logs, and managing log storage.

## Configuration Initialization
### InitConfig
- **Function**: Initializes the configuration by reading a specified YAML file.
- **Process**:
    1. **Read Configuration**: Uses Viper to read the YAML configuration file.
    2. **Unmarshal Configuration**: Converts the YAML configuration into the `Cfg`  struct.
    3. **Error Handling**: If the configuration cannot be read or parsed, the application logs a fatal error and stops.
### EnsureGeoIpInitialized
- **Function**: Ensures the GeoIP database is initialized.
- **Process**:
    1. **Set Manticore Connection Details**: Updates Manticore connection details from the configuration.
    2. **Open GeoIP Database**: Opens the GeoLite2-Country database.
    3. **Error Handling**: Logs an error if the database cannot be opened but allows the application to continue running.
### Schedule Cleanup Job
- **Function**: Schedules a cron job to clean up old tables in the Manticore search engine.
- **Process**:
    1. **Setup Cron Job**: Configures the cron job based on the rotation period specified in the configuration.
    2. **Error Handling**: This logs an error if the cron job cannot be scheduled and retry cleaning later.
## Log Rotation
### Cleanup Old Tables
- **Function**: Drops tables with a date suffix older than 30 days.
- **Process**:
    1. **Ensure Manticore Connection**: Verify the connection to the Manticore search engine.
    2. **Get Tables**: Retrieves the list of tables from Manticore and their creation dates.
    3. **Drop Tables**: Drop tables older than 30 days.
    4. **Error Handling**: Logs errors if tables cannot be dropped but continue to attempt to clean other tables.
### Extract Tables and Their Dates
- **Function**: Extract the date from the table name.
- **Process**:
    1. **Split Table Name**: Extracts the date part from the table name.
    2. **Parse Date**: Converts the extracted string to a date object.
    3. **Error Handling**: Logs a warning if the table name does not contain a valid date and skips that table.
## Server Initialization
### StartLogListener
- **Function**: Initializes and starts a syslog server based on the provided configuration.
- **Process**:
    1. **Setup Syslog Server**: Configures the Syslog server to use the appropriate protocol (UDP/TCP).
    2. **Boot Server**: Starts the server.
    3. **Initialize Queues**: Creates log and retry queues.
    4. **Start Workers**: Initiates worker processes for log handling.
    5. **Error Handling**: Logs errors if the server cannot start or the protocol is unsupported.
### Protocol Check
- **Function**: Determines the protocol for the syslog server (UDP/TCP).
- **Process**:
    1. **Check Protocol**: Validates if the protocol is either UDP or TCP.
    2. **Error Handling**: Logs an error and stops if the protocol is unsupported.
## Log Processing
### Start Workers
- **Function**: Starts worker processes to handle incoming logs.
- **Process**:
    1. **Initialize Workers**: Creates and starts worker processes based on configuration.
    2. **Batch Processor Initialization**: This starts the batch processor to handle log batches.
### ProcessLogWithGeo
- **Function**: Process logs to enrich them with GeoIP data.
- **Process**:
    1. **Unmarshal Log**: Converts the log string to a JSON object.
    2. **GeoIP Enrichment**: Adds GeoIP information to the log.
    3. **Marshal Log**: Converts the enriched log back to a JSON string.
    4. **Error Handling**: Logs errors if GeoIP data cannot be added or the log cannot be parsed.
### Determine Table Name
- **Function**: Determines the appropriate table name for storing the log.
- **Process**:
    1. **Apply Conditions**: Checks if log data matches specific conditions to determine the table name.
    2. **Append Date**: Appends the current date to the table name if configured.
    3. **Default Table Name**: Uses the default table name if no conditions are matched.
## Batch Processing
### Batch Processor
- **Function**: Handles the batching of logs for insertion into Manticore.
- **Process**:
    1. **Add Log to Batch**: Adds logs to the appropriate batch based on the table name.
    2. **Batch Size Check**: Checks if the batch size has been reached.
    3. **Insert Log Batch**: Insert the batch into Manticore.
    4. **Retry Queue**: Adds logs to the retry queue if insertion fails.
### Insert Log Batch
- **Function**: Inserts a batch of logs into the Manticore search engine.
- **Process**:
    1. **Ensure Manticore Connection**: Ensures a connection to Manticore.
    2. **Create Table**: Creates the table if it does not exist.
    3. **Insert Logs**: Executes the insert query for the batch.
    4. **Error Handling**: Adds logs to the retry queue if insertion fails.
## Manticore Processing
### Ensure Manticore Connection
- **Function**: Establishes and verifies a connection to the Manticore search engine.
- **Process**:
    1. **Initialize Client**: Initialize the Manticore client.
    2. **Update Index**: Updates the list of current indexes in Manticore.
    3. **Retry Mechanism**: Retries the connection multiple times if it fails.
    4. **Error Handling**: Logs errors if the connection cannot be established after multiple attempts.
### Create a Table if it Does Not Exist
- **Function**: Creates a table in Manticore if it does not already exist.
- **Process**:
    1. **Generate Create Query**: Creates the SQL query for table creation.
    2. **Execute Query**: Executes the create table query.
    3. **Error Handling**: Logs errors if the table cannot be created.
### Insert Logs
- **Function**: Inserts logs into the specified table in Manticore.
- **Process**:
    1. **Format Insert Query**: Formats the SQL insert query with the logs.
    2. **Execute Query**: Executes the insert query.
    3. **Error Handling**: Adds logs to the retry queue if insertion fails.
### Retry Queue
- **Function**: Processes logs in the retry queue.
- **Process**:
    1. **Wait for Interval**: Waits for a specified interval before retrying.
    2. **Retry Insertion**: Attempts to insert logs from the retry queue.
    3. **Error Handling**: Continues to retry until successful.
## Conclusion
The Log Listener application provides a comprehensive solution for receiving, processing, and storing log data. By leveraging GeoIP enrichment and robust error handling, the application ensures high reliability and efficiency in log management. Integrating with the Manticore search engine facilitates efficient storage and retrieval of logs, making it a powerful tool for log monitoring and analysis.

