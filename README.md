Product Data Processing with AWS Lambda
Overview
This project is a backend for an online store, managing product data with AWS Lambda. It has two Go functions: one loads products from a CSV file, and the other serves them via a REST API. I used PostgreSQL for storage, Redis for caching, and ensured it works locally and on AWS. It’s simple, efficient, and robust.
What I Built

uploadProduct: Reads products.csv, validates data, saves to PostgreSQL, and caches in Redis.
getAllProducts: Serves /products API, fetching from Redis or PostgreSQL.

My Thought Process

Database: PostgreSQL with UUIDs and timestamps for unique IDs and tracking.
uploadProduct: Processes CSV, supports local/S3, upserts data.
getAllProducts: Prioritizes Redis, falls back to PostgreSQL.
Local Testing: Docker and SAM for smooth development.
AWS: Scalable with RDS and ElastiCache.

How to Add Products

Edit products.csv:
Add rows, e.g.:name,image,price,qty
Laptop,http://example.com/laptop.jpg,999.99,10
Phone,,499.99,0


Save in project root.


Copy products.csv:Copy-Item products.csv .aws-sam\build\UploadProductFunction\


Clear Data:psql -h localhost -U postgres -d product_db -c "TRUNCATE TABLE products;"
docker exec -it mobile-cloud-lab_redis redis-cli flushall


Run uploadProduct:sam local invoke UploadProductFunction -e events\s3_event.json --env-vars env.json --add-host=host.docker.internal:host-gateway



How to Run Locally

Start Docker:docker-compose up -d


Initialize Database:psql -h localhost -U postgres -d product_db -f schema.sql

Password: root
Install Dependencies:go get github.com/lib/pq github.com/redis/go-redis/v9 github.com/joho/godotenv github.com/aws/aws-lambda-go
go mod tidy


Copy products.csv:Copy-Item products.csv .aws-sam\build\UploadProductFunction\


Build:sam build


Run uploadProduct:sam local invoke UploadProductFunction -e events\s3_event.json --env-vars env.json --add-host=host.docker.internal:host-gateway


Run getAllProducts:sam local start-api --env-vars env.json --add-host=host.docker.internal:host-gateway
Invoke-RestMethod http://127.0.0.1:3000/products


Verify Data:
Database: psql -h localhost -U postgres -d product_db -c "SELECT * FROM products;"
Redis: docker exec -it mobile-cloud-lab_redis redis-cli get "products:all"



Challenges and Fixes

Redis (nil): Redis key products:all was empty because uploadProduct failed to run. Ensured products.csv was in .aws-sam\build\UploadProductFunction\ and verified database/Redis connections.
Empty ID/Timestamps: API showed empty id and 0001-01-01 timestamps. Fixed by querying the database in uploadProduct to capture UUIDs and timestamps for Redis.
File Not Found: uploadProduct failed due to missing products.csv. Fixed by copying the file to the SAM build directory.
Compilation Errors: Resolved undefined: db.DBClient by creating db/db.go and adding database/sql import.

What I Learned

File Placement: Correctly placing products.csv in the SAM build directory is critical.
Caching: Ensuring database-generated fields are cached properly fixes data issues.
Debugging: Logs and direct database/Redis checks are key to troubleshooting.

This project demonstrates my ability to build and debug a serverless system, resolving issues like caching and file errors.


here are all the commmands ::
Start from Scratch

Stop Docker containers:
docker-compose down


Delete SAM build directory:
Remove-Item -Recurse -Force "C:\Users\Backb\Desktop\New folder (2)\mobile-cloud-lab\.aws-sam"


Start Docker containers:
docker-compose up -d


Recreate PostgreSQL database:
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS product_db;"
psql -h localhost -U postgres -c "CREATE DATABASE product_db;"
psql -h localhost -U postgres -d product_db -f schema.sql


Clear Redis:
docker exec -it mobile-cloud-lab_redis redis-cli FLUSHALL


Build SAM application:
sam build


Copy products.csv:
Copy-Item "C:\Users\Backb\Desktop\New folder (2)\mobile-cloud-lab\products.csv" "C:\Users\Backb\Desktop\New folder (2)\mobile-cloud-lab\.aws-sam\build\UploadProductFunction\"


Run uploadProduct:
sam local invoke UploadProductFunction -e events\s3_event.json --env-vars env.json --add-host=host.docker.internal:host-gateway


Verify Redis:
docker exec -it mobile-cloud-lab_redis redis-cli
GET products:all


Verify PostgreSQL:
psql -h localhost -U postgres -d product_db -c "SELECT * FROM products;"


Run getAllProducts:
sam local start-api --env-vars env.json --add-host=host.docker.internal:host-gateway


Test API:
Invoke-RestMethod -Uri http://127.0.0.1:3000/products
