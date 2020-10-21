# SimpleBank

A `go` service a simple bank. It has APIs for the frontend to do following things:

- Create and manage bank accounts, which are composed of owner’s name, balance, and currency.
- Record all balance changes to each of the account. So every time some money is added to or subtracted from the account, an account entry record will be created.
- Perform a money transfer between 2 accounts. This should happen within a transaction, so that either both accounts’ balance are updated successfully or none of them are.


## References

- [Simple Bank Tutorial](https://dev.to/techschoolguru/design-db-schema-and-generate-sql-code-with-dbdiagram-io-4ko5)
- [Design a SQL database schema](dbdiagram.io)