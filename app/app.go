//go:build mage
// +build mage

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	_ "github.com/lib/pq"
	"github.com/magefile/mage/mg"
)

type App mg.Namespace

// ConnectionString outputs a connection string for the database from env vars
func (App) ConnectionString() error {
	pgSSLMode := "require"
	map1 := map[string]string{
		"PGUSER":     os.Getenv("PGUSER"),
		"PGHOST":     os.Getenv("PGHOST"),
		"PGDATABASE": os.Getenv("PGDATABASE"),
		"PGPASSWORD": os.Getenv("PGPASSWORD"),
		"PGSSLMODE":  pgSSLMode,
	}
	for k, v := range map1 {
		if v == "" {
			return fmt.Errorf("missing environment variable %s", k)
		}
	}
	connStr := fmt.Sprintf(
		"host=%s port=5432 dbname=%s user=%s password=%s sslmode=%s",
		map1["PGHOST"],
		map1["PGDATABASE"],
		map1["PGUSER"],
		map1["PGPASSWORD"],
		map1["PGSSLMODE"],
	)
	fmt.Printf(connStr)
	return nil
}

// Token gets a token using `azidentity.NewDefaultAzureCredential`
func (App) Token() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	opts := policy.TokenRequestOptions{
		Scopes: []string{
			"https://ossrdbms-aad.database.windows.net/.default",
		},
	}

	ctx := context.Background()
	token, err := cred.GetToken(ctx, opts)
	if err != nil {
		return err
	}
	fmt.Printf(token.Token)
	return nil
}

// Serve runs a web server for our application
func (App) Serve() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, time.Now().Format(time.RFC3339))
	})

	pg := PG{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpRandom(w, r, pg)
	})

	mux.HandleFunc("/image", httpImage)

	listenAddr := ":8080"
	fmt.Printf("Listening on %s\n", listenAddr)
	err := http.ListenAndServe(listenAddr, httpLog(mux))
	return err
}

func httpLog(handler http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Info(
				"httpLog",
				"remoteAddr", r.RemoteAddr,
				"method", r.Method,
				"url", r.URL.String(),
			)
			handler.ServeHTTP(w, r)
		})
}

func httpRandom(w http.ResponseWriter, r *http.Request, pg PG) {
	db, err := pg.Open()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
	SELECT id, product_type_id, supplier_id, sku, name, price, description, image, digital, 
		   unit_description, package_dimensions, weight_in_pounds, reorder_amount, status, 
		   requires_shipping, warehouse_location, created_at, updated_at 
	FROM products 
	ORDER BY RANDOM() LIMIT 1
	`

	product := &Product{}
	row := db.QueryRow(query)
	err = row.Scan(
		&product.ID, &product.ProductTypeID, &product.SupplierID, &product.SKU, &product.Name,
		&product.Price, &product.Description, &product.Image, &product.Digital,
		&product.UnitDescription, &product.PackageDimensions, &product.WeightInPounds,
		&product.ReorderAmount, &product.Status, &product.RequiresShipping,
		&product.WarehouseLocation, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Error fetching product", http.StatusInternalServerError)
		return
	}

	// output product as indented JSON
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(product)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

func httpImage(w http.ResponseWriter, r *http.Request) {
	//return an error that says not implemented
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

type Product struct {
	ID                int       `json:"id"`
	ProductTypeID     int       `json:"product_type_id"`
	SupplierID        int       `json:"supplier_id"`
	SKU               string    `json:"sku"`
	Name              string    `json:"name"`
	Price             float64   `json:"price"`
	Description       string    `json:"description,omitempty"`
	Image             string    `json:"image,omitempty"`
	Digital           bool      `json:"digital"`
	UnitDescription   string    `json:"unit_description,omitempty"`
	PackageDimensions string    `json:"package_dimensions,omitempty"`
	WeightInPounds    string    `json:"weight_in_pounds,omitempty"`
	ReorderAmount     int       `json:"reorder_amount"`
	Status            string    `json:"status"`
	RequiresShipping  bool      `json:"requires_shipping"`
	WarehouseLocation string    `json:"warehouse_location,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Ping pings the database
func (App) Ping() error {
	pg := PG{}
	db, err := pg.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// ping the database
	err = db.Ping()
	if err != nil {
		return err
	}
	fmt.Printf("Connection successful!\n")
	return nil
}

// Tables lists the tables in the database
func (App) Tables() error {
	pg := PG{}
	db, err := pg.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return err
	}
	defer rows.Close()

	// print the tables to stdout
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return err
		}
		fmt.Println(tableName)
	}
	return nil
}

type PG struct {
	token *azcore.AccessToken
}

func (d *PG) Open() (*sql.DB, error) {
	connString, err := d.connString()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *PG) connString() (string, error) {
	if d.token == nil ||
		d.token.ExpiresOn.Sub(time.Now()) < 5*time.Minute {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return "", err
		}
		ctx := context.Background()
		opts := policy.TokenRequestOptions{
			Scopes: []string{
				"https://ossrdbms-aad.database.windows.net/.default",
			},
		}
		token, err := cred.GetToken(ctx, opts)
		if err != nil {
			return "", err
		}
		d.token = &token
	}

	pgSSLMode := "require"
	map1 := map[string]string{
		"PGUSER":     os.Getenv("PGUSER"),
		"PGHOST":     os.Getenv("PGHOST"),
		"PGDATABASE": os.Getenv("PGDATABASE"),
		"PGPASSWORD": d.token.Token,
		"PGSSLMODE":  pgSSLMode,
	}
	for k, v := range map1 {
		if v == "" {
			return "", fmt.Errorf("missing environment variable %s", k)
		}
	}
	connStr := fmt.Sprintf(
		"host=%s port=5432 dbname=%s user=%s password=%s sslmode=%s",
		map1["PGHOST"],
		map1["PGDATABASE"],
		map1["PGUSER"],
		map1["PGPASSWORD"],
		map1["PGSSLMODE"],
	)
	return connStr, nil
}
