# Additional Setup

This guide covers some optional configuration steps to get the best experience from Retro AIM Server.

## Configure User Directory Keywords

AIM users can make themselves searchable by interest in the user directory by configuring up to 5 interest keywords. The
keywords are organized by category.

Retro AIM Server does not come with any keywords installed out of the box. The following steps explain how to add
keywords and keyword categories via the management API.

1. **Add Keyword Categories**

   The following requests add 3 keyword categories:

    ```shell
    curl -d'{"name": "Programming Languages"}' http://localhost:8080/directory/category
    curl -d'{"name": "Books"}' http://localhost:8080/directory/category
    curl -d'{"name": "Music"}' http://localhost:8080/directory/category
    ```

   The following request lists all keyword categories along with their IDs, which are required for associating keywords
   to categories.

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

2. **Add Keywords**

   Two types of keywords are supported: categorized keywords, which belong to a specific category (e.g., Books, Music),
   and top-level keywords, which appear at the top of the menu and are not associated with any category.

   The following request sets up 3 keywords for books, music, and programming languages categories, respectively.

    ```shell
    curl -d'{"category_id": 2, "name": "The Dictionary"}' http://localhost:8080/directory/keyword
    curl -d'{"category_id": 3, "name": "Rock"}' http://localhost:8080/directory/keyword
    curl -d'{"category_id": 1, "name": "golang"}' http://localhost:8080/directory/keyword
    ```

   This request adds a single top-level keyword.

    ```shell
    curl -d'{"name": "Live, laugh, love!"}' http://localhost:8080/directory/keyword
    ```

   Fully rendered, the keyword list looks like this in the AIN client:

    <p align="center">
        <img width="500" alt="screenshot of AIM interests keywords menu" src="https://github.com/user-attachments/assets/f5295867-b74e-4566-879f-dfd81b2aab08">
    </p>

3. **Restart**

   After creating or modifying keyword categories and keywords, users currently connected to the server must sign out
   and back in again in order to see the updated keyword list.