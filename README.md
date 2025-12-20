# Price service
Sometimes people and business need to analyze the product's info in different markets to find the best price for the products. 

This often requires a lot of time when you 
compares prices from every market. With this service you can forget about this routine work and simple automate it.

<hr>

**in-brief:** *This service provides the API for getting the filtered samples of products in the most popular country's markets.*
<hr>
<hr>

## Main description
Here you can find the common description of the service's API.

The API-service was made as simple as possible for fast remembering the API-paths and parameters.

#### Note
It's recommended to use this service only into the **protected perimeter** (as local microservice ideally) 
because it was developed as *the internal local service* that doesn't have the global protection from external web-attacks.

*This service is the server-side of the user service app that lets users to visualize the result of products analysis.*

### API

#### Query parameters

##### Сlassification

- `necessary_parameter`

  this type of parameters **must be set** with the request where it uses.
  
  If it doesn't set *the error will return in the response from service*.
  
  <hr>


- `extra-parameter_with_default_value`

  this type of parameters **unrestrictly can be set** with the request where it uses.

  If it doesn't set *the default value will use*.
  
  <hr>

##### Parameters

Every request can be accompanied with the next query parameters that can be the extra or necessary that depends on the API-path you use. 

Classification can make search simplier. 

If you set the wrong value of the parameters the service will use the *default value* if the parameters' logic doesn't prescribe other.

- `query` :  `necessary_parameter`

  this parameter is a **required** query parameter that **must be set**. It formates according to the common URL-encoding.

  It's recommended to set this parameter as exact as possible because the response's accuracy will depend on it.

  <hr>

- `markets` : `necessary_parameter`

  this parameters is a **required** query parameter that **must be set**. It uses the next template:

  `{market_1}%20{market_2}%20{market_3}...` where {market_i} is the value that defines what market service must parse:
  - `wildberries`
  - `megamarket`
 
  <hr>

- `sample` : `extra-parameter_with_default_value`

  this parameter defines the number of products' sample that you want to get.

  It can be compared with the number of products' page in the markets.

  **Default value:** `1`.

  It must be a natural number.

  <hr>

- `amount` : `extra-parameter_with_default_value`

  this parameter defines the amount of the products that you want to get from the definite `sample`. It can be equal `min` or `max`.

  **Default value:** `min`.

  Every real amount depends on the definite market.

  If you use `min` value:

  - ***Wildberries***: 15 products (or less if the products doesn't exist in this amount).
  - ***MegaMarket***: 15 products (or less if the products doesn't exist in this amount).

  If you use the `max` value:

  - ***Wildberries***: `44` products (or less if the products doesn't exist in this amount).
  - ***MegaMarket***: `100` products (or less if the products doesn't exist in this amount).

  <hr>
 
- `sort` : `extra-parameter_with_default_value`

  this parameter defines the sort products' sort function.

  **Default value:** `popular`

  It can be equal the next types of values:
  - `popular`: sorts the sample by the popular.
  - `pricedown`: sorts the sample by the decreasing price range.
  - `priceup`: sorts the sample by the increasing price range.
  - `newly`: sorts the sample by the newest products.
 
  <hr>
 
- `no-image` : `extra-parameter_with_default_value`

  this parameter defines the extra-parsing way.

  It can be equal the `1` or `0` as the `true` and `false` respectively.

  If it set in `1` the product's image links won't be parsed.

  If it set in `0` the product's image links will be parsed.

  This parameter can optimize getting the products because of the reducing the extra-network calls.  

  **Default-value:** `1`

  This parameter influences only on the some markets that parse with using the browser's driver. The next markets' parsers use it while the products' getting:
  - ***Wildberries*** 

  Other parsers *don't use it and ignore it*.

  <hr>

- `price_down` : `necessary_parameter`

  this parameters defines the lower bound of the price range.

  It must be equal one of the `{0, 1, 2, ...}`

  <hr>

- `price_up` : `necessary_parameter`

  this parameters defines the upper bound of the price range.

  It must be equal one of the `{0, 1, 2, ...}` and *bigger than* `price_down`

  <hr>


#### Paths

The API has the next paths that you can use for getting the products with your requirments.

Make a note that paths contains here only the necessary params.

You can specified it as you want with the **extra-parameter_with_default_value**.

- `/products/filter/price/price-range?query={your_query}&sample={num}&price_down={lower_bound}&price_up={upper_bound}&markets={market_1}%20{market_2}%20...`
  
  this API-path provides the calls for getting the products filtered by the set price range.

  `[GET]`

  <hr>

- `/products/filter/price/best-price?query={your_query}&sample={num}&markets={market_1}%20{market_2}%20...`

  this API-path provides the calls for getting the products with the minimum price.

  `[GET]`

  <hr>

- `/products/filter/price/exact-price?query={your_query}&sample={num}&price={exact_price}&markets={market_1}%20{market_2}%20...`

  this API-path provides the calls for getting the products with the closest to `{exact_price}` price. The answer won't differ on more than **5% from exact-price**

  Parameter `price` : `necessary_parameter`.

  `price` can be equal the values like `price_down` and `price_up` values.

  `[GET]`

  <hr>

- `/products/filter/markets?query={your_query}&sample={num}&markets={market_1}%20{market_2}%20...`

  this API-path provides the calls for getting the products from the markets without any specified filter.

  `[GET]`

  <hr>

- `/products/filter/price/best-price/async?query={your_query}&sample={num}&markets={market_1}%20{market_2}%20...`

  this API-path provides the calls for getting the products with the minimum price in the async mode through the Kafka.

  `[POST]`

  If you need to have the extra-headers (for identitification, for example) you have to add the body to your request.

  The response on the request is in the next view:

  - Topic of response: `products`
  - Value: `JSON-object of chttp.ProductResponse`
  - Extra headers from the POST-request's body

  <hr>

- `/api/markets`

  this API-path provides the calls for getting the current markets that are supported by this service.

  `[GET]`

  <hr>


#### P.S.
For more information about the API see the ***swagger-API-docs*** using the endpoint `/swagger/index.html`

## How to install
This service is a part of the **best-price-project**. It must starts with the other services.

For a simpler installation and setup process, refer to the [best-price-project-deployment](https://github.com/MaKcm14/best-price-project-deployment) guide.

But if you need to use this service **independently** you must configure the kafka's cluster and start it using the docker compose.

### Configuring the .env file:
At the root directory you can find .env file that sets the default settings of this service. Here you can find some description about the .env file's params:

```
SOCKET="your_socket_that_will_use_for_starting_this_service"
BY_PASS_SOCKET="localhost:9090"
BROKERS="your_kafka_brokers'_sockets_divided_by_space_(bootstrap_list)"
ASYNC_MODE={"no" for unavailable accepting the async-requests || "yes" otherwise}
```
You can customize it.

#### Note:
The `.env_ui` file at the root has the default settings **specially for using with the price-service UI**.

Please *rename* it to the `.env` before using in this mode!

## Technology stack

- Echo Go Framework
- Flask Python Framework
- Docker
- Swagger
- Kafka
- Unit-Testing

## P.S.
This service is the main microservice of the best-price-project.

See the [price-service-tg-bot](https://github.com/MaKcm14/price-service-tg-bot) to find more information about the UI for this service.

See the [best-price-project-deployment](https://github.com/MaKcm14/best-price-project-deployment) to find more information about how to configure this service with the UI service.
