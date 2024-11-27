# golang-auth

# Run the following command to build the Docker image and start the container:
```
docker-compose up --build
```
# signup endpoints: 
User type can be ADMIN or USER
```
curl -X POST http://localhost:8000/users/singup \
-H "Content-Type: application/json" \
-d '{
    "first_name": "userName",
    "last_name": "userLastname",
    "password": "jelloworld",
    "phone": "0000000000",
    "user_type": "ADMIN",
    "email": "admin@email.com"
}'
```

# To check if you have access to user documnet or not
```
curl -X GET http://localhost:8000/users/user \
-H "Content-Type: application/json" \
-H "token: put your token"
```

# signin command
```
curl -X POST http://localhost:8000/users/signin \
-H "Content-Type: application/json" \
-d '{
    "email": "admin@email.co",
    "password": "jelloworld"
}'
```

# token revocation
```
curl -X POST http://localhost:8000/users/revokeToken \
-H "Content-Type: application/json" \
-H "token: your token" \
-d '{
    "email": "user email you want to revoke",
    "token": "user token"
}'

```

# get New Access token
```
curl -X POST http://localhost:8000/users/refreshToken \
-H "Content-Type: application/json" \
-d '{
    "refresh_token": "refresh token"
}'

```
