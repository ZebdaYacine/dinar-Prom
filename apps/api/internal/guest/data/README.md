# Algeria location catalogue

`algeria_58_wilayas.json` contains the 58-wilaya structure and all 1,541
communes used by the public guest catalogue. It is embedded into the API so
the location selector works offline and produces the same result in development
and production.

Source: [wadiemendja/algeria-communes-wilayas](https://github.com/wadiemendja/algeria-communes-wilayas)

The source repository publishes the data under the MIT license. The project
currently keeps the 58-wilaya structure because that is the structure defined
by the Dinar DZ PRD and public API contract; the 69-wilaya reform can be added
as a versioned catalogue later without changing the selector component.
