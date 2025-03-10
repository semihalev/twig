package main

import (
	"fmt"
	"os"

	"github.com/semihalev/twig"
)

func main() {
	// Create a new Twig engine
	engine := twig.New()

	// Create a template with set tag (using single quotes for string literals)
	simpleTemplate := `
<html>
<body>
    <h1>{{ shop_name }}</h1>
    
    {% set greeting = 'Welcome to our shop!' %}
    <p>{{ greeting }}</p>
    
    {% set year = 2025 %}
    <p>Copyright {{ year }}</p>
    
    {% set company = shop_name %}
    <p>{{ company }}</p>
    
    {% set message = greeting ~ ' Come visit us!' %}
    <p>{{ message }}</p>
    
    {% set price = 100 %}
    <p>Original price: ${{ price }}</p>
    
    {% set discount = 20 %}
    {% set final_price = price - discount %}
    <p>Final price after ${{ discount }} discount: ${{ final_price }}</p>
    
    <h2>Products</h2>
    {% set discount_rate = 15 %}
    <p>Special discount: {{ discount_rate }}%</p>
    
    <ul>
    {% for product in products %}
        {% set discount_amount = product.price * 0.15 %}
        {% set sale_price = product.price - discount_amount %}
        <li>
            {{ product.name }} - Original: ${{ product.price }}, Sale: ${{ sale_price }}, You save: ${{ discount_amount }}
        </li>
    {% endfor %}
    </ul>
</body>
</html>
`

	// Register the template
	err := engine.RegisterString("shop_template", simpleTemplate)
	if err != nil {
		fmt.Printf("Error registering template: %v\n", err)
		return
	}

	// Create a context with some products
	context := map[string]interface{}{
		"shop_name": "Twig Marketplace",
		"products": []map[string]interface{}{
			{
				"name":  "Laptop",
				"price": 1200,
			},
			{
				"name":  "Phone",
				"price": 800,
			},
			{
				"name":  "Headphones",
				"price": 200,
			},
		},
	}

	// Render the template
	fmt.Println("Rendering complex shop template with set tags:")
	err = engine.RenderTo(os.Stdout, "shop_template", context)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}
}
