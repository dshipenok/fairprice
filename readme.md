# Fair price app

"Online index price" generation made in several simple steps:

1. Combine data from different sources (channels) into single channel with price values.
2. Use the channel as input for "fair price" processor which gathers data and generates
   final price at the end of each period. Separate constant defines period's duration.
3. Logic of "fair price" could be easily replaced. Right now implemented "latest" and "average" strategies.

To be able to run application random data generators were used.
Period of data generation was set to 5s, just not to get bored :)

How to run:

```cli
go run cmd/cmg.go
```

## Classes

`pkg.Multiplexor`: combines channels into single one. Controls error channels as well.

`pkg.FairPrice`: processes data from single channel and put them into collector.

`pkg.collector.Average`: generates average price of each period (looks not so fair).

`pkg.collector.Latest`: generates latest price within each period.

`pkg.MockRandomStream`: fake random price generator.

Don't know what to write else.
