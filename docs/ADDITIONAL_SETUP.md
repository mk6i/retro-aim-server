# Additional Setup

This guide covers some optional configuration steps to get the best experience from Retro AIM Server.

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