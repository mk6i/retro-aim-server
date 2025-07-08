# Additional Setup

This guide covers some optional configuration steps to get the best experience from Retro AIM Server.

- [Configure User Directory Keywords](#configure-user-directory-keywords)
- [Configure SSL for Connecting AIM 6.2+](#configure-ssl-for-connecting-aim-62)

## Configure User Directory Keywords

AIM users can make themselves searchable by interest in the user directory by configuring up to 5 interest keywords.

Two types of keywords are supported: categorized keywords, which belong to a specific category (e.g., Books, Music), and
top-level keywords, which appear at the top of the menu and are not associated with any category.

Retro AIM Server does not come with any keywords installed out of the box. The following steps explain how to add
keywords and keyword categories via the management API.

1. **Add Categories**

   ###### Windows PowerShell

   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8080/directory/category" `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"name": "Programming Languages"}'

   Invoke-WebRequest -Uri "http://localhost:8080/directory/category" `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"name": "Books"}'

   Invoke-WebRequest -Uri "http://localhost:8080/directory/category" `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"name": "Music"}'
   ```

   ###### macOS / Linux / FreeBSD

    ```shell
    curl -d'{"name": "Programming Languages"}' http://localhost:8080/directory/category
    curl -d'{"name": "Books"}' http://localhost:8080/directory/category
    curl -d'{"name": "Music"}' http://localhost:8080/directory/category
    ```

2. **List Categories**

   Retrieve a list of all keyword categories created in the previous step.

   ###### Windows PowerShell

   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8080/directory/category" -Method GET
   ```

   ###### macOS / Linux / FreeBSD

    ```shell
    curl http://localhost:8080/directory/category
    ```

   This output shows the categories and their corresponding IDs, which you will use to assign keywords in the next step.

    ```json
    [
      {
        "id": 2,
        "name": "Books"
      },
      {
        "id": 3,
        "name": "Music"
      },
      {
        "id": 1,
        "name": "Programming Languages"
      }
    ]
    ```

3. **Add Keywords**

   The first 3 requests set up keywords for books, music, and programming languages categories using the category IDs
   from the previous step. This last request adds a single top-level keyword with no category ID.

   ###### Windows PowerShell

   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8080/directory/keyword" `
      -Method POST `
      -ContentType "application/json" `
      -Body '{"category_id": 2, "name": "The Dictionary"}'

   Invoke-WebRequest -Uri "http://localhost:8080/directory/keyword" `
      -Method POST `
      -ContentType "application/json" `
      -Body '{"category_id": 3, "name": "Rock"}'

   Invoke-WebRequest -Uri "http://localhost:8080/directory/keyword" `
      -Method POST `
      -ContentType "application/json" `
      -Body '{"category_id": 1, "name": "golang"}'

   Invoke-WebRequest -Uri "http://localhost:8080/directory/keyword" `
      -Method POST `
      -ContentType "application/json" `
      -Body '{"name": "Live, laugh, love!"}'
   ```

   ###### macOS / Linux / FreeBSD

    ```shell
    curl -d'{"category_id": 2, "name": "The Dictionary"}' http://localhost:8080/directory/keyword
    curl -d'{"category_id": 3, "name": "Rock"}' http://localhost:8080/directory/keyword
    curl -d'{"category_id": 1, "name": "golang"}' http://localhost:8080/directory/keyword
    curl -d'{"name": "Live, laugh, love!"}' http://localhost:8080/directory/keyword
    ```

   Fully rendered, the keyword list looks like this in the AIM client:

    <p align="center">
        <img width="500" alt="screenshot of AIM interests keywords menu" src="https://github.com/user-attachments/assets/f5295867-b74e-4566-879f-dfd81b2aab08">
    </p>

   Check out the [API Spec](../api.yml) for more details on directory API endpoints.

4. **Restart**

   After creating or modifying keyword categories and keywords, users currently connected to the server must sign out
   and back in again in order to see the updated keyword list.

## Configure SSL for Connecting AIM 6.2+

AIM 6.2 and later clients require SSL to authenticate via Kerberos. To support this, you'll need to run an SSL
termination proxy that forwards traffic to Retro AIM Server over plain TCP.

### Overview

This project provides:

- A Docker-based toolchain to build OpenSSL and stunnel
- Scripts to generate self-signed certificates
- Tools to create and load a certificate into a certificate database compatible with AIM 6.x clients

> **Note:** AIM 6.x clients use a legacy SSLv2-style `ClientHello`, which is not supported by modern OpenSSL. This
> project includes a Docker build of the last OpenSSL version that works.

### What Is the NSS Certificate Database?

AIM 6.x clients validate SSL certificates using
an [NSS (Network Security Services)](https://developer.mozilla.org/en-US/docs/Mozilla/Projects/NSS) certificate
database, which stores trusted root certificates.

You'll need to distribute a certificate database containing your root certificate to any client machines running AIM
6.x, which do not use the system CA store and will only trust certificates present in their NSS certificate
database. Even if your certificate is signed by a public certificate authority, you must explicitly add it to the NSS
database used by the client.

### Prerequisites

- [Docker](https://docs.docker.com/get-started/get-docker/) installed and running
- A Unix-like terminal (for `make` and shell scripts)

### Setup Steps

These steps will generate a self-signed SSL certificate (if needed), configure an NSS certificate database for AIM
clients, and launch an SSL termination proxy using stunnel.

#### 1. Clone the repository

Clone [Retro AIM Server](https://github.com/mk6i/retro-aim-server) and run the following commands from the root of the
repository.

#### 2. Build Docker images

This sets up containers used for certificate generation and SSL proxying:

```sh
make stunnel-image cert-gen-image
```

#### 3. Generate certificates and certificate database

> Replace `ras.dev` with your chosen domain.

```sh
make certs CERT_NAME=ras.dev
```

This will:

- Generate a self-signed certificate for the provided domain.
- Save certs to the `certs/` directory.
- Create an NSS certificate database at `certs/nssdb` and add the cert to it.

If you're not using a real domain you own, add an entry to your `/etc/hosts` file pointing the domain to your server's IP.

#### 4. Run stunnel (SSL terminator)

> Replace `ras.dev` with your chosen domain.

```sh
./scripts/run_stunnel.sh ./certs/ras.dev.pem
```

This launches `stunnel`, which:

- Accepts SSL connections from AIM clients on port 443.
- Decrypts the traffic.
- Forwards plain TCP traffic to Retro AIM Server.

To customize ports, certificate paths, or backend forwarding settings, edit:

```
config/ssl/stunnel.conf
```