# Fenix API with sub-routes

Using sub-routes system with negroni and gorilla to generate auth if needed.

```
acct := acctBase.PathPrefix("/account").Subrouter()
acct.Path("/movies").HandlerFunc(CreateMovieEndPoint).Methods("POST")
```

/account os the sub-route that requires jwt verification and /movies is the regular route

###  Generate RSA signing files via shell (adjust as needed):

$ openssl genrsa -out app.rsa 1024

$ openssl rsa -in app.rsa -pubout > app.rsa.pub