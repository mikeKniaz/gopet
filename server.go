package main

import(
  "io/ioutil"
  "log"
  "fmt"
  "time"
  "strconv"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/render"
  "github.com/martini-contrib/binding"
  "database/sql"
  _ "github.com/lib/pq"
  "net/http"
)

const TouristHost = "http://localhost:3000/api/v2"

// export PATH=$PATH:/usr/local/go/bin
// curl -H "Content-Type: application/json" -X POST -d '{"id":1,"name":"Test1","kind":"Tour"}' http://localhost:8000/products
// curl -i localhost:8000/products

func SetupDB() *sql.DB {
  db, err := sql.Open("postgres", "user=gopg password=gopgpass dbname=goproducts sslmode=disable")
  PanicIf(err)
  return db
}

func PanicIf(err error) {
  if err != nil {
    panic(err)
  }
}

func main() {
  m := martini.Classic()

  m.Map(SetupDB())
  m.Use(render.Renderer(render.Options{
    Charset: "UTF-8",
  }))

  m.Get("/", func() string {
    return "Golang pet, hello!"
  })

  m.Get("/products/external/:id", func(params martini.Params) string {
    id, _ := strconv.Atoi(params["id"])
    return getTintProduct(id)
  })

  m.Post("/products", binding.Bind(Product{}), func(product Product, r render.Render) {
    p := Product { product.Id, product.Name, product.Kind, time.Now().Format("2006-01-02 15:04:05") }

    r.JSON(201, p.save())
  })

  m.Get("/products", func(params martini.Params, r render.Render) {
    r.JSON(200, allProducts())
  })

  m.Get("/products/:id", func(params martini.Params, r render.Render) {
    id, _ := strconv.Atoi(params["id"])
    r.JSON(200, findProduct(id))
  })

  m.RunOnAddr(":8000")
}

// external request
func getTintProduct(id int) string {
  productUrl := fmt.Sprintf("%s/%s/%d", TouristHost, "products", id)

  client := &http.Client{}
  request, error := http.NewRequest("GET", productUrl, nil)
  request.SetBasicAuth("react_test", "9d8d059e0a649b2b5f724cfedfd3c1d7")

  response, error := client.Do(request)
  if error != nil {
    log.Fatalln(error)
  }
  body, error := ioutil.ReadAll(response.Body)
  return string(body)
}

// product struct
type Product struct {
  Id int
  Name, Kind, CreatedAt string
}

func findProduct(id int) Product {
  db := SetupDB()
  rows, err := db.Query("SELECT * FROM products WHERE id = $1 LIMIT 1", id)
  PanicIf(err)

  products := []Product{}
  for rows.Next() {
    PanicIf(rows.Err())
    p := Product{}
    err := rows.Scan(&p.Id, &p.Name, &p.Kind, &p.CreatedAt)
    PanicIf(err)
    products = append(products, p)
  }
  return products[0]
}

func allProducts() []Product {
  db := SetupDB()
  rows, err := db.Query("SELECT * FROM products")
  PanicIf(err)

  products := []Product{}
  for rows.Next() {
    PanicIf(rows.Err())
    p := Product{}
    err := rows.Scan(&p.Id, &p.Name, &p.Kind, &p.CreatedAt)
    PanicIf(err)
    products = append(products, p)
  }
  return products
 }

func (p Product) save() Product {
  db := SetupDB()
  _, err := db.Query("INSERT INTO products VALUES ($1, $2, $3, $4)", p.Id, p.Name, p.Kind, p.CreatedAt)
  PanicIf(err)
  return p
}
