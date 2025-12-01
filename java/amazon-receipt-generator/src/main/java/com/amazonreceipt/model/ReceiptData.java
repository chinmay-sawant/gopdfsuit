package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;
import java.util.List;

/**
 * Root data model for the receipt JSON.
 */
public class ReceiptData {
    @SerializedName("config")
    private ReceiptConfig config;

    @SerializedName("title")
    private TitleSection title;

    @SerializedName("table")
    private List<TableSection> tables;

    @SerializedName("image")
    private List<Object> images;

    @SerializedName("footer")
    private FooterSection footer;

    public ReceiptConfig getConfig() {
        return config;
    }

    public void setConfig(ReceiptConfig config) {
        this.config = config;
    }

    public TitleSection getTitle() {
        return title;
    }

    public void setTitle(TitleSection title) {
        this.title = title;
    }

    public List<TableSection> getTables() {
        return tables;
    }

    public void setTables(List<TableSection> tables) {
        this.tables = tables;
    }

    public List<Object> getImages() {
        return images;
    }

    public void setImages(List<Object> images) {
        this.images = images;
    }

    public FooterSection getFooter() {
        return footer;
    }

    public void setFooter(FooterSection footer) {
        this.footer = footer;
    }
}
