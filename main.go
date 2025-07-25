package main

import (
	goload "goload/runner"
	"time"
)

func main() {
	config := goload.Config{
		Request: goload.Request{
			Method:    goload.GET,
			URI:       "http://localhost:8974/products/100",
			UserAgent: goload.ChromeAgent,
		},
		Response: goload.Response{
			StatusCode: 200,
			Body:       "{\"productID\":100,\"user\":{\"userID\":77,\"email\":\"oumaima@gmail.com\",\"firstName\":\"Mark\",\"lastName\":\"Charles\",\"address\":null,\"phoneNumber\":null,\"profilePicture\":\"26b67ab4-4e5b-42a6-8e41-0727508fb1ad.png\",\"gender\":null,\"city\":null,\"verified\":false,\"createdAt\":null,\"rating\":{\"rating\":5.0,\"numberOfRatings\":1}},\"category\":{\"categoryId\":0,\"name\":null,\"cover\":null},\"cities\":[{\"cityId\":20,\"countyCode\":\"CA\",\"name\":\"Moose Jaw\"}],\"sizes\":[{\"id\":59,\"categoryId\":8,\"sizeName\":\"XS\"}],\"materials\":[{\"id\":48,\"categoryId\":8,\"material\":\"Cotton\"}],\"brand\":null,\"name\":\"YSL Scarf 2\",\"description\":\"ede145\",\"condition\":\"like new\",\"price\":145.0,\"thumbnail\":\"197b70fbb5166ce7ce41995a540587ef-2910057047003213315-thumbnail.jpg\",\"genders\":[\"male\",\"female\"],\"images\":[\"197b70fbb5166ce7ce41995a540587ef-2910057047003213315.jpg\"],\"colors\":[\"ff03a9f4\",\"ffff5722\",\"ffffffff\"],\"favorite\":false,\"appContact\":true,\"negotiable\":false,\"phoneNumber\":null,\"active\":true,\"processing\":false,\"createdAt\":\"2024-12-15\"}",
		},
		LogOutputPath: "logs",
		Timepoints: []goload.ExecutionTimepoint{
			{
				Duration: time.Second * 2,
				TargetVu: 10,
			},
		},
	}

	duration, err := time.ParseDuration("30s")
	if err != nil {
		panic(err)
	}

	config.Timeout = duration
	goload.Execute(config)
}
