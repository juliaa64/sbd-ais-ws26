# Exercise 2 — Operation Data Plane with OLTPs

**Goal:** run analytical queries on a single-node OLTP database (PostgreSQL),
find out *where the time actually goes* when a query does not scale, and
discuss which fixes — query, hardware, or architecture — actually help.

## Part 0 — Before the session

The PostgreSQL image is ~450 MB. Downloading it over the classroom Wi-Fi is
slow when everybody does it at once, so please do this beforehand:

```bash
docker compose pull
```

Then check that:

1. `docker ps` shows the running container `pg-bigdata` after do `docker compose up -d`.
2. `python3 --version` works on your machine (on Windows use `python` or `py`
   instead of `python3` in all commands below).

You need ~2 GB of free disk space. See [Troubleshooting](#troubleshooting) if something fails.

## Part 1 — Environment Setup and Basics (guided walkthrough, not graded)

### 1. Start the environment

If it is not running already:

```bash
docker compose up -d
```

### 2. Access PostgreSQL

```bash
docker exec -it pg-bigdata psql -U postgres
```

### 3. Load and query data in PostgreSQL

#### 3.1 Create a large dataset (in another terminal)

```bash
cd data
python3 expand.py
```

Creates `data/people_1M.csv` with 1 million rows (takes a few seconds).

```bash
wc -l people_1M.csv
```

#### 3.2 Enter PostgreSQL

```bash
docker exec -it pg-bigdata psql -U postgres
```

#### 3.3 Create and load the table

```sql
DROP TABLE IF EXISTS people_big;

CREATE TABLE people_big (
  id SERIAL PRIMARY KEY,
  first_name TEXT,
  last_name TEXT,
  gender TEXT,
  department TEXT,
  salary INTEGER,
  country TEXT
);

\COPY people_big(first_name,last_name,gender,department,salary,country) FROM '/data/people_1M.csv' DELIMITER ',' CSV HEADER;
```

#### 3.4 Enable timing

```sql
\timing on
```

### 4. Verification

```sql
SELECT COUNT(*) FROM people_big;
SELECT * FROM people_big LIMIT 10;
```

### 5. Examples of analytical queries

#### (a) Simple aggregation

```sql
SELECT department, AVG(salary)
FROM people_big
GROUP BY department
LIMIT 10;
```

#### (b) Nested aggregation

```sql
SELECT country, AVG(avg_salary)
FROM (
  SELECT country, department, AVG(salary) AS avg_salary
  FROM people_big
  GROUP BY country, department
) sub
GROUP BY country
LIMIT 10;
```

#### (c) Top-N sort

```sql
SELECT *
FROM people_big
ORDER BY salary DESC
LIMIT 10;
```

> **Think about it:** does `LIMIT 10` make queries (a) and (b) cheaper?
> Compare with and without it using `EXPLAIN ANALYZE`.
> 
> ANSWER: 
> Query (a) costs 17189.60..17192.64 and (b) 18811.57..18938.78 using `LIMIT 10`.
> If we don't use it, the costs are 17189.60..17215.44 and 20538.48..20540.98 respectively.
> So there is not a lot of difference between (a) with limit and without, but with it is a bit better.
> And in the case of (b), the costs are less similar and also the result with limit is better.

## Part 2 — Activities (graded)

### Activity 2.1 — PostgreSQL Analytical Queries (E-commerce)

Generate the dataset (in another terminal, from the `ecommerce` folder):

```bash
cd ecommerce
python3 dataset_generator.py
```

This writes `data/orders_1M.csv`, which is available inside the PostgreSQL
container as `/data/orders_1M.csv`. Its columns are:

```
customer_name, product_category, quantity, price_per_unit, order_date, country
```

Load the generated data into PostgreSQL in a **new table** called `orders`.
Write the `CREATE TABLE` yourself (choose sensible types for each column) and
use the same `\COPY` pattern as in section 3.3.

ANSWER:
```
DROP TABLE IF EXISTS orders;

CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  customer_name TEXT,
  product_category TEXT,
  quantity INTEGER,
  price_per_unit FLOAT,
  order_date DATE,
  country TEXT
);

\COPY orders(customer_name, product_category, quantity, price_per_unit, order_date, country) FROM '/data/orders_1M.csv' DELIMITER ',' CSV HEADER;
```

Using SQL ([list of supported SQL commands](https://www.postgresql.org/docs/current/sql-commands.html)),
answer the following questions:

**A.** Which order has the highest `price_per_unit`?
>841292 | Emma Brown    | Automotive       |        3 |           2000 | 2024-10-11 | Italy
```
SELECT *
FROM orders
ORDER BY price_per_unit DESC
LIMIT 1;
```

**B.** What are the top 3 product categories with the highest total quantity
sold across all orders?
>Health & Beauty (300842), Electronics (300804) and Toys (300598)
```
SELECT product_category, SUM(quantity) AS sum_quantity
FROM orders
GROUP BY product_category
ORDER BY sum_quantity DESC
LIMIT 3;
```

**C.** What is the total revenue per product category?
(Revenue = `price_per_unit × quantity`)
> * Automotive       | 306589798.86000013
> * Books            | 12731976.040000021
> * Electronics      | 241525009.44999996
> * Fashion          | 31566368.220000133
> * Grocery          | 15268355.660000077
> * Health & Beauty  |  46599817.89000009
> * Home & Garden    |  78023780.09000018
> * Office Supplies  |  38276061.63999993
> * Sports           |  61848990.83000004
> * Toys             | 23271039.020000055
```
SELECT product_category, SUM(price_per_unit * quantity) AS revenue
FROM orders
GROUP BY product_category;
```

**D.** Who are the top 5 customers by total spending?
> * **customer_name  |  total_spending   | count**
> * Carol Taylor   | 991179.1800000002 |  1028
> * Nina Lopez     | 975444.9499999998 |   980
> * Daniel Jackson |         959344.48 |  1033
> * Carol Lewis    | 947708.5700000001 |   943
> * Daniel Young   | 946030.1400000001 |   973
``` 
SELECT customer_name, SUM(price_per_unit * quantity) AS total_spending, COUNT(*)
FROM orders
GROUP BY customer_name
ORDER BY total_spending DESC
LIMIT 5;
```

**E.** Look at the spending totals in D — and at how many orders each of those
customers has. What do you notice? Open `ecommerce/dataset_generator.py` and
explain *why* the data looks like this.

> All of them have similar numbers in total spending and also in how many orders 
> they have. If we check `ecommerce/dataset_generator.py`, we can find that the
> functions choose the data randomly (the costumer and the category). This means
> that the results are going to be uniform and there are not going to be outliers.

### Activity 2.2 — Why Is This Self-Join So Slow?

Users of the system sometimes run naive queries such as:

```sql
SELECT COUNT(*)
FROM people_big p1
JOIN people_big p2
  ON p1.country = p2.country;
```

On the full 1M-row table this takes **more than 10 minutes** and slows down
the whole database for everyone else. Your job is to find out *why* before
proposing a fix.

> **Do not run it on the full table in class** — use the smaller tables below.
> (If you are curious, run it at home and let it finish.)

**Step 1 — Measure how it grows.** Create three smaller copies of the table:

```sql
CREATE TABLE people_50k  AS SELECT * FROM people_big WHERE id <= 50000;
CREATE TABLE people_100k AS SELECT * FROM people_big WHERE id <= 100000;
CREATE TABLE people_200k AS SELECT * FROM people_big WHERE id <= 200000;
```

Run the self-join (with `\timing on`) on each of the three tables and fill in:

| rows in table | join result (`COUNT(*)`) | time |
|---|---|---|
| 50 000 | 27501822| 3095.555 ms (00:03.096)|
| 100 000 | 109946508| 14935.895 ms (00:14.936)|
| 200 000 | 439395606| 80464.699 ms (01:20.465)|

```sql
SELECT COUNT(*)
FROM people_50k p1
JOIN people_50k p2
  ON p1.country = p2.country;
```

When the input **doubles**, by what factor do the result and the time grow?
Use this to **predict** the result size and the runtime on `people_big` (1M rows).
> The query exhibits quadratic growth (O(N^2)). When the input size doubles 
> (e.g., from 50k to 100k), the result size and the execution time grow by a 
> factor of 4 (2^2).
> 
> Since 1 million is 5 times larger than 200,000, the result size and time will 
> grow by a factor of 25 (5^2). If the 200k query takes around 1 minute and
> 20 seconds (80464.699 ms), the 1M row query will take 25 * 80464.699 ms =
> 2011617,475 ms (over 9 hours and a half).

**Step 2 — Does an index help?** Create an index on `country` of `people_100k`,
run `ANALYZE people_100k;`, and repeat the query. Did the time change? Use
`EXPLAIN ANALYZE` to support your answer. Why does (or doesn't) the index help?

```sql
CREATE INDEX index_people_100k_country ON people_100k (country);
ANALYZE people_100k;
        
EXPLAIN ANALYZE
SELECT COUNT(*)
FROM people_100k p1
JOIN people_100k p2
ON p1.country = p2.country;
```
> The execution time was 26434.683 ms. 
> 
> No, an index on country does not significantly improve the execution time. 
> `EXPLAIN ANALYZE` will show that the database still has to perform a massive 
> join operation (the execution time was longer).

**Step 3 — Rewrite it.** The query only wants the *number* of matching pairs,
not the pairs themselves. If a country has *k* people, how many pairs does it
contribute to the join? Write a query that computes the **same number
without a join**. Check that it returns exactly the same result as the join on
`people_100k`, then run it on `people_big` and compare its runtime with your
prediction from Step 1.
> If a country has k people, the self-join creates k pairs of each person of 
> this country (k*k). So we can calculate the same if we calculate the k^2 of
> each country.

```sql
SELECT SUM(k * k)
FROM (
    SELECT COUNT(*) AS k
    FROM people_100k
    GROUP BY country
);
```
> It returns 109946508, the same number as the self-join. If we run it on
> people_big, we obtain 10983941260 and a runtime of 178.722 ms. 

**Step 4 — Discussion (submit in writing).** Considering **scalability** and
**efficiency**, which approaches and/or optimizations can be applied to improve
this kind of query in a real system? Discuss at least:

- what the rewrite in Step 3 tells you about adding more hardware or an index;
- what you would do if the business actually needed **the pairs themselves**
  (not just their count) — would a bigger machine or a cluster help, and how
  much?
- the limits of an **OLTP database** for this workload, especially in a
  **large-scale cloud environment**.

> * The rewrite in Step 3 demonstrates that fixing algorithmic complexity 
> (moving from O(N^2) to O(N)) is more effective than scaling hardware or 
> adding indexes.
> 
> * If the pairs themselves is required, a single database node will struggle 
> with memory. Adding a bigger machine (vertical scaling) helps slightly by 
> providing more RAM and being faster.
> 
> * OLTP databases (like PostgreSQL) are built for quick, small tasks. 
> They are not meant to do heavy math on millions of rows at once. If you run 
> a massive query (like the self-join) on an OLTP, it will use up all the 
> server's power (CPU and memory). This will freeze the database and lock 
> out normal users trying to use your app. 

> **Optional:** support your answer with a diagram, SQL or code.

## Submission

Commit to the Github repository and add the link to the Github solution in the Moodle submission:

- **2.1:** your `CREATE TABLE` statement and the SQL for A–D with their
  results, and your answer to E;
- **2.2:** the table from Step 1 with your prediction, your observations from
  Steps 2–3 (including your rewrite query), and the written discussion from
  Step 4.

Be ready to shortly present your solutions (5–8 minutes) in the next exercise
session (05.11.2026).

## Clean up

```bash
docker compose down
```

This removes the container, **including the PostgreSQL tables** — the CSV
files in `data/` are kept.

Optionally, delete the generated datasets as well (~95 MB). They are
not needed any more, and running `expand.py` and `dataset_generator.py`
again recreates exactly the same files:

```bash
rm data/people_1M.csv data/orders_1M.csv
```

(Windows PowerShell: `Remove-Item data/people_1M.csv, data/orders_1M.csv`)

## Troubleshooting

- **`port is already allocated` / `address already in use`** — another
  PostgreSQL (port 5432) is running on your machine.
  Stop it, or change the left-hand port in `docker-compose.yml`
  (e.g. `"5433:5432"`).
- **`python3: command not found` (Windows)** — use `python` or `py`.
- **`\COPY ... No such file or directory`** — the CSV was not generated into
  the `data/` folder. Check that `data/people_1M.csv` / `data/orders_1M.csv`
  exist on your machine.
