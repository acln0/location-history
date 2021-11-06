## Location History REST service
This is my submission to the Flink go programming challenge.
### Routes
- `POST` `/v1/location/history/{order_id}`
- `GET` `/v1/location/history/{order_id}?max=<N>`
- `DELETE` `/v1/location/history/{order_id}`

### Issues/Discussion points
- Issues due to time constraints
  - Lack of tests
  - Lack of mocks
  - API documentation

- #### TTL
    The original submission contained an erroneous condition which rendered the location history TTL useless.
    I've fixed it in this repo, and kept the original erroneous condition as a comment on `store/history_store.go:31`
    
    My approach for the bonus task (History location TTL) was to have a separate go routine that would continuously check the `lastAccessed`
    property on every Order's location history and delete it if it hasn't been accessed for longer thant he configured TTL.
    
    One issue with this, is the way `lastAccessed` is set. With the current implementation it would only update when Creating or Updating an Order's location history.
    
    Meaning even if a location history is constantly being fetched it will still be deleted (this is bad i know :D). A quick and easy fix would be to also update the `lastAccessed` property on GET requests.
    
    An alternative approach would be some sort of lazy TTL, where instead of constantly checking if items are expired. We would check if items are expired only when trying to retrieve or append to the Location history. Coupled with a CRON job, we can also delete stale items that are not accessed frequently. 
    
- #### routing
  Since this repo doesn't depend on any external libraries. The routing is very rudimentary and has all the inconveniences that comes with the go standard library, such as retrieving path params.
  That's why the pathPrefix is forwarded to the Handler so that it's easier to trim and only modify in one place in case routes change.