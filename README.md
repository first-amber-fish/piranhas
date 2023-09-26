<p align="center">
<img src="piranha.png" width="150">
</p>

piranhas 
==============================

Piranhas is a collection of tasks to work with exported and non-exported fields of structs.
Enabling stuctures with defaults values using [struct tags](http://golang.org/pkg/reflect/#StructTag).

Installation
------------

The recommended way to install piranhas

```
go get github.com/first-amber-fish/piranhas
```

Default
--------

Default deals with setting the default values of a struct. For this the struct tag key 'default' is evaluated and set independently of the previous value of field. 

The value of the 'default' tag key should match the data type. The typical formats are specified here, as these are used by the parse functions belonging to the data type. For the datatypes complex64 and complex128 an own parser was developed, which expects cartesian data in the format 'a+bi'. For time the standard golang parser is used. This still needs a layout, which led to it, it beside the 'default' tag key still the 'layout' tag key is queried. If no layout tag key is defined, 'RFC3339' is used. There are a lot of predefined time layouts in golang. You can use them here or define your own layout.  Both a type specification 'RFC822' or '02 Jan 06 15:04 MST' are valid.

Structures, arrays, slices and maps are recursively passed through and the function for setting defaults is called on them. Empty arrays and slices are ignored, as well as nil pointers. 

Upper and lower case of field names, as if the variable is exported or not, does not matter.

Examples
--------

A basic example:

```go
import (
    "fmt"
    "time"

    "github.com/first-amber-fish/piranhas"
)

...

type structMapExample struct {
    floatVar float64 `default:"87.5"`
}

type embeddedExample struct {
    numberVar int `default:"1234"`
}

type structExample truct {
    numberVar int `default:"5678"`
}

type example struct {
    boolVar        bool           `default:"true"` //<-- StructTag with a default key
    stringVar1     string         `default:"33"`
    stringVar2     *string        `default:"33"`
    nixVar         int8
    durVar         time.Duration  `default:"1m"`
    dateVar        time.Time      `default:"04.09.1990" layout:"02.01.2006"` //<-- StructTag with a default and layout key
    complexVar     complex128     `default:"3.5+2.7i"`
    stringSlice    []string       `default:"[\"a\",\"b\"]"`
    stringMapOfInt map[string]int `default:"{\"a\": 5,\"b\": 6}"`
    sliceVar       []structExample
    mapVar         map[string]structExample
    structVar1     structExample
    structVar2     *structExample
    embeddedExample
}

...

var exampleVar example

piranhas.SetDefaults(&exampleVar) //<-- This set the defaults values

fmt.Println(exampleVar.boolVar)              //Prints: true
fmt.Println(exampleVar.stringVar2)           //Prints: 33
fmt.Println(exampleVar.nixVar)               //Prints: 0
fmt.Println(exampleVar.durVar)               //Prints: 1m0s
fmt.Println(exampleVar.numberVar)            //Prints: 1234
fmt.Println(exampleVar.structVar1.numberVar) //Prints: 5678
fmt.Println(exampleVar.stringSlice[0])       //Prints: a
fmt.Println(exampleVar.stringMapOfInt["a"])  //Prints: 5

...
```

Path
----

Path determines the value of simple variables in complex structures. Especially when structs, slices and maps are an instance of documents, it can be challenging to determine the correct value in a pre-programmed way. Path works similar to a file path on the operating system, which uses different directories as location information for a file. Path refers here not to directories and file, but to structs, array, slices and maps and fields as location indication and returns simple fields. Path is not a query language, but specifies in a point notation the way to get to the required field. 

If a path element hits a struct, the path element is interpreted as a field name. If it is a slice or array, it is interpreted as an index, where 0 is the first index element. If it is a map, the path element is interpreted as a key.  

Path normally does'nt return whole structs, slices, or maps, but only scalar data types. Exceptions are the data types []Byte (ByteSlice), Time and Duration. Returned is always a copy of the field value, so that the original struct can't be changed. Upper and lower case of fields, as if the variable is exported or not, does not matter.

The function GetPathInterface returns the result as interface{}. The user can now examine the data type and then convert it to the target type as needed with a type assertion. For easier use, for each data type returned there is a special function, GetPathDataType(), which takes over this task and returns the correct data type. 

A small challenge is the path as text. This can have both a '.' dot as separator, as well as a '/' slash or an '\' backslash. But also a '[' is understood as a separator.  

Examples
--------
A basic path transformation example:

```
$..foo.bar       ->  foo.bar
$.baz[0].qux     ->  baz.0.qux
$.baz.[0].qux    ->  baz.0.qux
baz["0"][v].qux  ->  baz.0.v.qux
baz/0/v/qux      ->  baz.0.v.qux

```

A basic code example:

```go
import (
    "fmt"
    "time"
    
	"github.com/first-amber-fish/piranhas"
)

...

type address struct {
	street string
	number int
	city   string
	ZIP    string
}

type passport struct {
	number string
}

type person struct {
	passport
	firstName            string
	lastName             *string
	age                  int
	developer            bool
	address              address
	adresses1            []address
	hobbys               map[string]int
	fingerprint          []byte
	birthDate            time.Time
	concentrationAbility time.Duration
}

...

lastName := "Ranseier"
data := person{
	passport: passport{"KI123"},
	firstName: "Karl",
	lastName:  &lastName,
	age:       58,
	developer: true,
	address: address{
		street: "Tellerstraße",
		number: 29,
		city:   "Berlin",
		ZIP:    "10553",
	},
	adresses1: []address{
		{
			street: "Müllerstr",
			number: 129,
			city:   "Berlin",
			ZIP:    "10487",
		},
		{
			street: "Kanzlerplatz",
			number: 1,
			city:   "Berlin",
			ZIP:    "10000",
		},
	},
	hobbys:               map[string]int{"Motorcycle": 10, "Skydiving": 9, "Crochet": 0},
	fingerprint:          []byte{72, 101, 108, 108, 111},
	birthDate:            time.Date(1965, time.June, 9, 3, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
	concentrationAbility: 2*time.Hour + 35*time.Minute,
}

...

piranhas.GetPathString(&data,"number")                 // returns KI123
piranhas.GetPathDuration(&data,"concentrationAbility") // returns 2h35m
piranhas.TestGetPathTime(&data, "birthDate")           // returns 9 of June 1965
piranhas.TestGetPathByteSlice(&data,"fingerprint")     // returns []byte{72, 101, 108, 108, 111}
piranhas.GetPathString(&data,"lastName")               // returns Ranseier
piranhas.GetPathBool(&data,"developer")                // returns true
piranhas.GetPathString(&data,"address.city")           // returns Berlin
piranhas.GetPathString(&data,"address.street")         // returns Tellerstraße
piranhas.GetPathString(&data,"adresses1.0.street")     // returns Müllerstr
piranhas.GetPathString(&data,"adresses1.1.ZIP")        // returns 10000
piranhas.GetPathInt(&data,"hobbys.Motorcycle")         // returns 10

...
```

License
-------

MIT, see [LICENSE](LICENSE)
