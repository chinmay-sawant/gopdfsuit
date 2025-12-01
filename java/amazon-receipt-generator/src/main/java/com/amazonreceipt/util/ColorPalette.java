package com.amazonreceipt.util;

import com.itextpdf.kernel.colors.Color;
import com.itextpdf.kernel.colors.DeviceRgb;

/**
 * Amazon-inspired color palette for beautiful PDF styling.
 */
public class ColorPalette {
    
    // Amazon Brand Colors
    public static final Color AMAZON_ORANGE = new DeviceRgb(255, 153, 0);
    public static final Color AMAZON_DARK = new DeviceRgb(19, 25, 33);
    public static final Color AMAZON_BLUE = new DeviceRgb(0, 113, 133);
    
    // Primary Colors
    public static final Color PRIMARY_DARK = new DeviceRgb(35, 47, 62);
    public static final Color PRIMARY_LIGHT = new DeviceRgb(72, 89, 112);
    
    // Background Colors
    public static final Color BG_LIGHT_GRAY = new DeviceRgb(248, 249, 250);
    public static final Color BG_MEDIUM_GRAY = new DeviceRgb(237, 240, 242);
    public static final Color BG_WHITE = new DeviceRgb(255, 255, 255);
    
    // Text Colors
    public static final Color TEXT_PRIMARY = new DeviceRgb(15, 17, 17);
    public static final Color TEXT_SECONDARY = new DeviceRgb(86, 89, 89);
    public static final Color TEXT_MUTED = new DeviceRgb(118, 118, 118);
    
    // Border Colors
    public static final Color BORDER_LIGHT = new DeviceRgb(221, 221, 221);
    public static final Color BORDER_DARK = new DeviceRgb(204, 204, 204);
    
    // Accent Colors
    public static final Color ACCENT_GREEN = new DeviceRgb(0, 123, 58);
    public static final Color ACCENT_RED = new DeviceRgb(204, 12, 57);
    
    // Header Background
    public static final Color HEADER_BG = new DeviceRgb(35, 47, 62);
    public static final Color SECTION_HEADER_BG = new DeviceRgb(237, 240, 242);
    
    private ColorPalette() {
        // Utility class - prevent instantiation
    }
}
