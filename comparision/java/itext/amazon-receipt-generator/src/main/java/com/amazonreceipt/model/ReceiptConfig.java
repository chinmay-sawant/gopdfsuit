package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;

/**
 * Configuration settings for the PDF receipt.
 */
public class ReceiptConfig {
    @SerializedName("pageBorder")
    private String pageBorder;

    @SerializedName("page")
    private String page;

    @SerializedName("pageAlignment")
    private int pageAlignment;

    @SerializedName("watermark")
    private String watermark;

    public String getPageBorder() {
        return pageBorder;
    }

    public void setPageBorder(String pageBorder) {
        this.pageBorder = pageBorder;
    }

    public String getPage() {
        return page;
    }

    public void setPage(String page) {
        this.page = page;
    }

    public int getPageAlignment() {
        return pageAlignment;
    }

    public void setPageAlignment(int pageAlignment) {
        this.pageAlignment = pageAlignment;
    }

    public String getWatermark() {
        return watermark;
    }

    public void setWatermark(String watermark) {
        this.watermark = watermark;
    }

    /**
     * Parse page border values (top:right:bottom:left).
     */
    public float[] getBorderValues() {
        if (pageBorder == null || pageBorder.isEmpty()) {
            return new float[]{1, 1, 1, 1};
        }
        String[] parts = pageBorder.split(":");
        float[] values = new float[4];
        for (int i = 0; i < 4 && i < parts.length; i++) {
            values[i] = Float.parseFloat(parts[i]);
        }
        return values;
    }
}
