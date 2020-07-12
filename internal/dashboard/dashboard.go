package dashboard

import (
	"errors"
	"fmt"
	"log"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/mijies/dashboard_generator/internal/account"
	"github.com/mijies/dashboard_generator/internal/config"
	"github.com/mijies/dashboard_generator/pkg/utils"
)

// commands and ttl_codes
type dashboard_component interface {
	iterable()	iterator
	getLength() int
	getComponentLabel(cfg config.Config) string
	loadData(cfg config.Config, acc *account.UserAccount)
	finalize()
	intoRows()
}

func Build(book_path string, acc *account.UserAccount) {
	d := dashboard{
		base_path:	book_path,
		acc:		acc,
		cfg:		config.NewConfig(),
	}
	d.load()
	d.build()
}

type dashboard struct {
	base_path	string
	cfg			config.Config // interface
	acc			*account.UserAccount
	book		*excelize.File
	commands	commands
	ttl_codes	ttl_codes
}

func(d *dashboard) load() {
	file, err := excelize.OpenFile(d.base_path)
    if err != nil {
        log.Fatal(err)
	}
	d.book = file

	d.commands.loadData(d.cfg, d.acc)
	d.commands.finalize()
	d.ttl_codes.loadData(d.cfg, d.acc)
	d.ttl_codes.finalize()
}

func(d *dashboard) build() {
	// create a new book
	time_format := d.cfg.GetTimeFormat()
	new_path := utils.AddTimestampToFilename(d.base_path, time_format, "xlsm")
	if err := d.book.SaveAs(new_path); err != nil {
        log.Fatal(err)
	}

	// swap with the new one
	d.base_path = new_path
	file, err := excelize.OpenFile(d.base_path)
    if err != nil {
        log.Fatal(err)
	}
	d.book = file

	// delete the exsisting macro sheet
	sheet_name := d.cfg.GetMacroSheetName()
	d.book.DeleteSheet(sheet_name)

	// copy the template sheet
	d.book.NewSheet(sheet_name)
	tmp_index := d.book.GetSheetIndex(d.cfg.GetMacroTmpSheetName())
	index  := d.book.GetSheetIndex(sheet_name)
	if err := d.book.CopySheet(tmp_index, index); err != nil {
        log.Fatal(err)
	}

	d.renderSheet(sheet_name, &d.commands)
	d.renderSheet(sheet_name, &d.ttl_codes)
	if err := d.book.Save(); err != nil {
        log.Fatal(err)
	}
}

func(d *dashboard) renderSheet(sheet_name string, comp dashboard_component) {
	label := comp.getComponentLabel(d.cfg)
	rowc  := d.commands.getLength() + 10 // row count to cover
	rows  := [2]int{1, rowc}
	cols  := [2]int{1, 5}
	rowi, err := d.locateRow(sheet_name, label, rows, cols)
	if err != nil {
		log.Fatal(err)
	}

	seed := int('A')
	itr  := comp.iterable()
	for itr.hasNext() {
		rowi++
		d.book.DuplicateRow(sheet_name, rowi)
		cols := itr.next()
		for i, v := range cols {
			axis := fmt.Sprintf("%s%d", string(seed + i), rowi)
			d.book.SetCellValue(sheet_name, axis, v)
		}
	}
}

func(d *dashboard) locateRow(sheet_name string, value string, rows [2]int, cols [2]int) (int, error) {
	if cols[1] > 27 {
		log.Fatal("column only supported upto Z")
	}

	seed := int('A' - 1)
	for col := cols[0]; col < cols[1]; col++ {
		for row  := rows[0]; row < rows[1]; row++ {
			axis := fmt.Sprintf("%s%d", string(seed + col), row)
			val, err := d.book.GetCellValue(sheet_name, axis)
			if err != nil {
				log.Fatal(err)
			}		
			if val == value {
				return row, nil
			}
		}	
	} 
	return -1, errors.New("no cell found with the value:" + value)
}
