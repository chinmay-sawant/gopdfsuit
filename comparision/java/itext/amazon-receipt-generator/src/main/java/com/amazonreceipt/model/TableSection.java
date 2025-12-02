package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;
import java.util.List;

/**
 * Table section containing rows of data.
 */
public class TableSection {
    @SerializedName("maxcolumns")
    private int maxColumns;

    @SerializedName("columnwidths")
    private List<Integer> columnWidths;

    @SerializedName("rows")
    private List<TableRow> rows;

    public int getMaxColumns() {
        return maxColumns;
    }

    public void setMaxColumns(int maxColumns) {
        this.maxColumns = maxColumns;
    }

    public List<Integer> getColumnWidths() {
        return columnWidths;
    }

    public void setColumnWidths(List<Integer> columnWidths) {
        this.columnWidths = columnWidths;
    }

    public List<TableRow> getRows() {
        return rows;
    }

    public void setRows(List<TableRow> rows) {
        this.rows = rows;
    }

    /**
     * Get column widths as float array for iText.
     */
    public float[] getColumnWidthsAsFloatArray() {
        if (columnWidths == null || columnWidths.isEmpty()) {
            float[] defaultWidths = new float[maxColumns];
            for (int i = 0; i < maxColumns; i++) {
                defaultWidths[i] = 1f;
            }
            return defaultWidths;
        }
        float[] widths = new float[columnWidths.size()];
        for (int i = 0; i < columnWidths.size(); i++) {
            widths[i] = columnWidths.get(i).floatValue();
        }
        return widths;
    }
}
