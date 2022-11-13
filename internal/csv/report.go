package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"os"
	"path/filepath"
)

type Builder struct {
	logger *logging.Logger
}

func NewBuilder(l *logging.Logger) *Builder {
	return &Builder{logger: l}
}

func (b *Builder) CreateReport(rows []model.ReportRow, year, month int) (string, error) {
	filePath := fmt.Sprintf("static/reports/%d_%d_report.csv", year, month)

	_, err := os.Stat("static/reports")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := filepath.Join(".", "static/reports")
			err = os.MkdirAll(p, os.ModePerm)
			if err != nil {
				return "", err
			}
		}
	}

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

func (b *Builder) IsCreated(year, month int) (bool, string, error) {
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
