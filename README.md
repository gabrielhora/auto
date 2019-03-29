# Auto

Simple automation server (jenkins with less features?)

---

- Single database server that works as the "master"
- A **Server** will connect to the database and register itself
- **Servers** provide a administration web interface
- **User** connect to any **Server** to create and configure **Jobs**
- **Jobs** can run on demand or on a **Schedule**
- **Jobs** can be started with a HTTP Post to a specific URL
- **Jobs** have access to the JSON data sent via HTTP Post
- **Jobs** have execution history with logs, execution time and result (os exit code) 
- **Jobs** can run on one or more **Servers**
- Leverage database locking mechanisms to avoid scheduling race conditions 
