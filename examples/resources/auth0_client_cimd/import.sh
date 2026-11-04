# Import an existing CIMD client by its client ID (usually prefixed with "tpc_").
# For a dynamically registered client, import instead of creating a second
# client for the same external_client_id URL.
terraform import auth0_client_cimd.my_mcp_agent "tpc_5FPpaVyZGSNRCBzTb2zURZ"
