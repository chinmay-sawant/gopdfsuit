package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;

/**
 * Title section of the receipt.
 */
public class TitleSection {
    @SerializedName("props")
    private String props;

    @SerializedName("text")
    private String text;

    public String getProps() {
        return props;
    }

    public void setProps(String props) {
        this.props = props;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }
}
