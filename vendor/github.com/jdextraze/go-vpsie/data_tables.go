package vpsie

type Offer struct {
	Id      string `json:"id"`
	Cpu     int    `json:"cpu"`
	Ram     int    `json:"ram"`
	Ssd     int    `json:"ssd"`
	Traffic int    `json:"traffic"`
	Price   int    `json:"price"`
}

type offersResponse struct {
	baseResponse
	Offers []Offer `json:"offers"`
}

func (c *client) GetOffers() ([]Offer, error) {
	out := offersResponse{}
	return out.Offers, c.doGetRequest("offers", &out)
}

type Datacenter struct {
	Id      string `json:"id"`
	Name    string `json:"dc_name"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type datacentersResponse struct {
	baseResponse
	Datacenters []Datacenter `json:"datacenters"`
}

func (c *client) GetDatacenters() ([]Datacenter, error) {
	out := datacentersResponse{}
	return out.Datacenters, c.doGetRequest("datacenters", &out)
}

type Image struct {
	Id       string `json:"id"`
	Name     string `json:"image_name"`
	Category string `json:"category"`
}

type imagesResponse struct {
	baseResponse
	Images []Image `json:"images"`
}

func (c *client) GetImages() ([]Image, error) {
	out := imagesResponse{}
	return out.Images, c.doGetRequest("images", &out)
}
