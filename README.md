# Fileblobs

Fileblobs is a web application for managing files in Azure Blob Storage. It provides a user-friendly interface for browsing, uploading, and downloading files stored in Azure Blob Storage containers.

## Features

- **Authentication System**
  - Local user authentication
  - OIDC integration support
  - Role-based access control (admin/user)
  
- **Storage Account Management**
  - View all configured storage accounts
  - Add new storage account connections
  - Edit existing storage account details
  - Select an active storage account for operations
  
- **File Operations**
  - Browse files and folders with hierarchical navigation
  - Upload single or multiple files
  - Download individual files
  - Download entire folders (as zip archives)
  - Download multiple selected files (as zip archives)
  - Search for files within the current directory
  
- **Web Interface**
  - Responsive design with Bootstrap
  - File type icons for better visualization
  - Breadcrumb navigation
  - File search functionality

## Prerequisites

- Go 1.21 or higher
- Azure Storage Account
- Docker (optional, for containerized deployment)

## Installation

### Local Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd fileblobs
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file in the root directory with the following configuration:
   ```env
   # Server configuration
   PORT=80

   # Default Azure Storage Account (optional, can be configured through UI)
   AZURE_STORAGE_ACCOUNT_NAME=youraccountname
   AZURE_STORAGE_ACCOUNT_KEY=youraccountkey
   AZURE_STORAGE_CONTAINER=yourcontainername
   
   # Authentication settings (optional for OIDC)
   AUTH_TYPE=local          # 'local' or 'oidc'
   OIDC_PROVIDER_URL=       # If using OIDC, provide the provider URL
   OIDC_CLIENT_ID=          # If using OIDC, provide the client ID
   OIDC_CLIENT_SECRET=      # If using OIDC, provide the client secret
   OIDC_REDIRECT_URL=       # If using OIDC, provide the redirect URL
   ```

4. Create the data directory:
   ```bash
   mkdir -p data
   ```

5. Build and run the application:
   ```bash
   go build -o fileblobs ./cmd
   ./fileblobs
   ```

### Docker Deployment

1. Build the Docker image:
   ```bash
   docker build -t fileblobs .
   ```

2. Run the container:
   ```bash
   docker run -p 80:80 \
     -e AZURE_STORAGE_ACCOUNT_NAME=youraccountname \
     -e AZURE_STORAGE_ACCOUNT_KEY=youraccountkey \
     -e AZURE_STORAGE_CONTAINER=yourcontainername \
     -v $(pwd)/data:/app/data \
     fileblobs
   ```

## Configuration

### Initial Setup

On first run, the application will create a default admin user if no users exist:
- Username: `admin`
- Password: `admin`

You should change this password immediately after first login.

### Storage Accounts

You can configure multiple storage accounts through the web interface after logging in:

1. Navigate to the "Storage" page
2. Click "Add Account"
3. Fill in the storage account details:
   - Name: A friendly name for the storage account
   - Description: Optional description
   - Account Name: The Azure Storage account name
   - Account Key: The Azure Storage account key
   - Container Name: The blob container name

### Authentication Modes

The application supports two authentication modes:

1. **Local Authentication** (default)
   - Users are stored locally in the `data/auth.json` file
   - Administrators can add/edit users through the web interface

2. **OIDC Authentication** 
   - Requires configuration of OIDC provider settings in the `.env` file
   - Users will be redirected to the provider for login

## Usage

### Web Interface

The web interface is accessible at `http://localhost:80` (or whatever port you configured).

1. **Login**
   - Use your username and password to log in
   - If using OIDC, you'll be redirected to your identity provider

2. **Navigate Storage Accounts**
   - Click "Storage" in the top-right corner to view/manage storage accounts
   - Select an account to work with by clicking "Select"

3. **Browse Files**
   - Files and folders are displayed in a list
   - Click on folders to navigate into them
   - Use the breadcrumb navigation to move back up the hierarchy

4. **Upload Files**
   - Click the "Upload" button
   - Select one or more files from your computer
   - Files will be uploaded to the current directory

5. **Download Files**
   - Click on a file to download it directly
   - Use "Download Selected" to download multiple files as a zip archive
   - Use "Download Folder" to download the current folder as a zip archive

6. **Search Files**
   - Use the search box to filter files in the current view

### API Usage

The application provides HTTP endpoints for programmatic access:

#### Authentication

```http
POST /login
Content-Type: application/x-www-form-urlencoded

username=user&password=pass
```

A successful login will set a session cookie that should be included in subsequent requests.

#### File Operations

1. **List Files**
   ```http
   GET /?prefix=path/to/folder
   ```

2. **Download File**
   ```http
   GET /download?path=path/to/file
   ```

3. **Download Multiple Files**
   ```http
   POST /download-multiple
   Content-Type: application/x-www-form-urlencoded

   paths=path/to/file1&paths=path/to/file2
   ```

4. **Download Folder**
   ```http
   GET /download-folder?prefix=path/to/folder
   ```

5. **Upload Files**
   ```http
   POST /upload
   Content-Type: multipart/form-data

   prefix=path/to/folder
   files=@file1.txt
   files=@file2.txt
   ```

## Security Considerations

- The application stores sensitive information like storage account keys
- Production deployments should:
  - Use HTTPS with a valid SSL certificate
  - Run behind a reverse proxy (like Nginx)
  - Configure proper firewall rules
  - Use strong passwords and consider OIDC for authentication
  - Regularly update dependencies

## Troubleshooting

### Common Issues

1. **Cannot connect to Azure Storage**
   - Verify account name, key, and container are correct
   - Check network connectivity to Azure
   - Ensure the container exists

2. **Upload failures**
   - Verify the user has write permissions to the storage account
   - Check file size limits (application limits uploads to 32MB per request)

3. **Authentication issues**
   - Check that the data directory is writable
   - Verify OIDC configuration if using external authentication

## License

[License Information]

## Contributing

[Contribution Guidelines]
