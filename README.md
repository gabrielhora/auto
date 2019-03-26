# Auto

Simple automation server.

---

#### Characteristics

- Single database server that works as the *master*

- A **Server** will connect to the database and register itself

- **Servers** provide a administration web interface

- **User** connect to any **Server** to create and configure **Jobs**

- **Jobs** can run on demand or on one or more **Schedules**

- **Jobs** can be started with a HTTP Post to a specific URL

- **Jobs** have access to the JSON data sent via HTTP Post

- **Jobs** have execution history with logs, execution time and result (os exit code) 

- **Jobs** can run on one or more **Servers** (but always in a single server)

---

#### Scheduling

Each **Server** runs a background job that pools the database server for new **Jobs** to execute.

Every random amount of milliseconds between 10.000 and 60.000 a **Server** will try to acquire a
lock on a semaphore and check for pending **Job** executions based on the following algorithm: 

- Start a database transaction

- Run the following query to acquire an exclusive lock with timeout of 5 seconds
  ```sql
  LOCK TABLE queue IN EXCLUSIVE MODE
  ```

- If the query failed, exit

- If succeeded, run the following command to get pending jobs:
  ```sql
  DELETE FROM queue
  WHERE id IN (
    SELECT q.id 
    FROM queue q
    INNER JOIN job j ON j.id = q.job_id
    WHERE 
    ( -- check if job is runnable in this server
      j.runnable_any = true 
      OR j.runnable_in = ANY('{<server ID>}'::bigint[])
    )
    AND q.date <= '<now in UTC>'
  )
  RETURNING queue.job_id
  ``` 
  The query will pop from the queue table all **Jobs** that are pending to be execute

- For each **Job** returned in the previous query, schedule it's next execution based on the
**Job's** *cron* expression
  ```sql
  INSERT INTO queue (job_id, date) VALUES ('job_id', 'calculated_date')
  ``` 

- Commit the transaction (this will release the lock)

- Execute each **Job** returned by the query in previous steps
