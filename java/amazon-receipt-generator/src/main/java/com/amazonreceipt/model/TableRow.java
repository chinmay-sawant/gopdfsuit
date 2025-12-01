package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;
import java.util.List;

/**
 * A row in a table section.
 */
public class TableRow {
    @SerializedName("row")
    private List<CellData> row;

    public List<CellData> getRow() {
        return row;
    }

    public void setRow(List<CellData> row) {
        this.row = row;
    }
}
