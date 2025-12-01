package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;

/**
 * Footer section of the receipt.
 */
public class FooterSection {
    @SerializedName("font")
    private String font;

    @SerializedName("text")
    private String text;

    public String getFont() {
        return font;
    }

    public void setFont(String font) {
        this.font = font;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }
}
