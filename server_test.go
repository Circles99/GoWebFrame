package GoWebFrame

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestTpl(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}

	s := NewHttpServer(ServerWithTemplateEngine(&TemplateEngine{T: tpl}))
	s.Get("/login", func(c *Context) {
		err = c.Reader("login.gohtml", nil)
		if err != nil {
			t.Fatal(err)
		}
	})

	if err := s.Start(":8082"); err != nil {
		t.Fatal(err)
	}
}

// Supplier 结构体用于表示每个供应商的信息
type Supplier struct {
	IsDefault   bool   `json:"is_default"`
	Price       string `json:"price"`
	Priority    string `json:"priority"`
	SupplierID  string `json:"supplier_id"`
	WarehouseID string `json:"warehouse_id"`
}

// Product 结构体用于表示产品的信息
type Product struct {
	BusinessDeveloperName       string     `json:"business_developer_name"`
	BuyerName                   string     `json:"buyer_name"`
	CategoryID                  string     `json:"category_id"`
	ChineseCustomsName          string     `json:"chinese_customs_name"`
	EnglishCustomsName          string     `json:"english_customs_name"`
	EnglishName                 string     `json:"english_name"`
	FeatureLabels               []string   `json:"feature_labels"`
	HSCode                      string     `json:"hs_code"`
	IsPublic                    string     `json:"is_public"`
	IsSupportPlaceOrder         string     `json:"is_support_place_order"`
	OrderItemID                 string     `json:"order_item_id"`
	PrintTemplateType           string     `json:"print_template_type"`
	ProductLine                 string     `json:"product_line"`
	ProductionFileProcessMethod string     `json:"production_file_process_method"`
	RawMaterialID               string     `json:"raw_material_id"`
	SupplierSelectMode          string     `json:"supplier_select_mode"`
	Suppliers                   []Supplier `json:"suppliers"`
	Title                       string     `json:"title"`
	TongToolCategoryID          string     `json:"tong_tool_category_id"`
	Type                        string     `json:"type"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 解析表单数据
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 读取表单数据
	var product Product

	fmt.Println(r.PostForm)
	b, e := io.ReadAll(r.Body)
	if e != nil {
		log.Print(e)
	}
	if err := mapFormToStruct(b, &product); err != nil {
		fmt.Println(r.PostForm)
		fmt.Println(err)
		http.Error(w, "Failed to map form to struct", http.StatusBadRequest)
		return
	}

	// 输出表单数据
	fmt.Printf("Received data: %+v\n", product)

	// 发送响应
	w.Write([]byte("Data received successfully"))
}

func mapFormToStruct(formData []byte, v interface{}) error {
	//params := make(map[string]interface{})
	//for k := range formData {
	//	fmt.Println(k)
	//	fmt.Println(formData.Get(k))
	//	params[k] = formData.Get(k)
	//}

	//requestData, err := json.Marshal(params)

	// 将表单数据转换为 JSON
	//jsonData, err := json.Marshal(formData)
	//if err != nil {
	//	return err
	//}

	// 将 JSON 解码到结构体
	err := json.Unmarshal(formData, v)
	if err != nil {
		return fmt.Errorf("xxxxxxxxxxxxxxxxxx%s", err.Error())
	}

	return nil
}

func TestKKK(t *testing.T) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8083", nil)
}
