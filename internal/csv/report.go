package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
)

type Builder struct {
	logger *logging.Logger
}

func NewBuilder(l *logging.Logger) *Builder {
	return &Builder{logger: l}
}

func (*Builder) CreateReport(rows []model.ReportRow, year, month int) (string, error) {
	filePath := fmt.Sprintf("static/reports/%d_%d_report.csv", year, month)

	csvFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return "", err
	}
	w := csv.NewWriter(csvFile)

	data := [][]string{{"service_name", "total_revenue"}}
	for _, val := range rows {
		row := []string{val.ServiceName, val.Cost}
		data = append(data, row)
	}

	err = w.WriteAll(data)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (*Builder) IsCreated(year, month int) (bool, string, error) {
	filePath := fmt.Sprintf("static/reports/%d_%d_report.csv", year, month)

	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}

	return true, filePath, nil

}
