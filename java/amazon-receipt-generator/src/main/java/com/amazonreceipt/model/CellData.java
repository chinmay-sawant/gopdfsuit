package com.amazonreceipt.model;

import com.google.gson.annotations.SerializedName;

/**
 * Cell data containing properties and text.
 */
public class CellData {
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

    /**
     * Parse cell properties.
     * Format: font:size:style:alignment:borderTop:borderRight:borderBottom:borderLeft
     * Style: 100 = bold, 010 = italic, 001 = underline
     */
    public CellProperties getParsedProperties() {
        return new CellProperties(props);
    }

    /**
     * Inner class to hold parsed cell properties.
     */
    public static class CellProperties {
        private String fontName = "font1";
        private float fontSize = 10f;
        private boolean bold = false;
        private boolean italic = false;
        private boolean underline = false;
        private String alignment = "left";
        private boolean borderTop = true;
        private boolean borderRight = true;
        private boolean borderBottom = true;
        private boolean borderLeft = true;

        public CellProperties(String props) {
            if (props == null || props.isEmpty()) {
                return;
            }

            String[] parts = props.split(":");
            if (parts.length >= 1) {
                fontName = parts[0];
            }
            if (parts.length >= 2) {
                try {
                    fontSize = Float.parseFloat(parts[1]);
                } catch (NumberFormatException e) {
                    fontSize = 10f;
                }
            }
            if (parts.length >= 3) {
                String style = parts[2];
                if (style.length() >= 3) {
                    bold = style.charAt(0) == '1';
                    italic = style.charAt(1) == '1';
                    underline = style.charAt(2) == '1';
                }
            }
            if (parts.length >= 4) {
                alignment = parts[3];
            }
            if (parts.length >= 5) {
                borderTop = "1".equals(parts[4]);
            }
            if (parts.length >= 6) {
                borderRight = "1".equals(parts[5]);
            }
            if (parts.length >= 7) {
                borderBottom = "1".equals(parts[6]);
            }
            if (parts.length >= 8) {
                borderLeft = "1".equals(parts[7]);
            }
        }

        public String getFontName() {
            return fontName;
        }

        public float getFontSize() {
            return fontSize;
        }

        public boolean isBold() {
            return bold;
        }

        public boolean isItalic() {
            return italic;
        }

        public boolean isUnderline() {
            return underline;
        }

        public String getAlignment() {
            return alignment;
        }

        public boolean hasBorderTop() {
            return borderTop;
        }

        public boolean hasBorderRight() {
            return borderRight;
        }

        public boolean hasBorderBottom() {
            return borderBottom;
        }

        public boolean hasBorderLeft() {
            return borderLeft;
        }
    }
}
