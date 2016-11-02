package main

type SerialPort struct {
	Path        string `json:"path"`
	VendorId    uint16 `json:"vendorId"`
	ProductId   uint16 `json:"productId"`
	DisplayName string `json:"displayName"`
}
