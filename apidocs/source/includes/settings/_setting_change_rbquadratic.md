# Setting change rbquadratic (rebalance quadratic)

## Create setting rbquadratic

```shell
curl -X POST "https://gateway.local/v3/setting-change-rbquadratic" \
-H 'Content-Type: application/json' \
-d '{
      "change_list": [
        {
          "type":"update_asset",
          "data":{
              "asset_id":3,
              "rebalance_quadratic":{
                "a":0.000001754386,
                "b":0.000001754386,
                "c":0.000001754386
              }
          }
        },
        {
          "type":"update_asset",
          "data":{
              "asset_id":5,
              "rebalance_quadratic":{
                "a":0.00000357143,
                "b":0.0002285714,
                "c":0.99976786
              }
          }
        },
        ...
      ]
    }'
```

> sample response

```json
{
  "id": 6,
  "success": true
}
```

### HTTP Request

`POST https://gateway.local/v3/setting-change-rbquadratic`

Params | Type | Required | Default | Description
------ | ---- | -------- | ------- | -----------
asset_id | uint64 | true | nil | ID of asset
a | float64 | false | nil | 
b | float64 | false | nil | 
c | float64 | false | nil | 
<aside class="notice">Write key is required</aside>

## Get pending setting rbquadratic


```shell
curl -X GET "https://gateway.local/v3/setting-change-rbquadratic"
```

> sample response

```json
{
  "data": [
    {
      "id": 6,
      "created": "2019-08-13T07:25:49.869418Z",
      "change_list": [
        {
          "type":"update_asset",
          "data":{
              "asset_id":3,
              "rebalance_quadratic":{
                "a":0.000001754386,
                "b":0.000001754386,
                "c":0.000001754386
              }
          }
        },
        {
          "type":"update_asset",
          "data":{
              "asset_id":5,
              "rebalance_quadratic":{
                "a":0.00000357143,
                "b":0.0002285714,
                "c":0.99976786
              }
          }
        },
        ...
      ]
    }
  ],
  "success": true
}
```

### HTTP Request

`GET https://gateway.local/v3/setting-change-rbquadratic`
<aside class="notice">All keys are accepted</aside>

## Confirm pending setting rbquadratic

```shell
curl -X PUT "https://gateway.local/v3/setting-change-rbquadratic/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`PUT https://gateway.local/v3/setting-change-rbquadratic/:change_id`
<aside class="notice">Confirm key is required</aside>

## Reject pending setting rbquadratic

```shell
curl -X DELETE "https://gateway.local/v3/setting-change-rbquadratic/6"
```

> sample response

```json
{
  "success": true
}
```

### HTTP Request

`DELETE https://gateway.local/v3/setting-change-rbquadratic/:change_id`
<aside class="notice">Confirm key is required</aside>