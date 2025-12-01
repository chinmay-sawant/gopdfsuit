package com.amazonreceipt.util;

import com.itextpdf.kernel.font.PdfFont;
import com.itextpdf.kernel.font.PdfFontFactory;
import com.itextpdf.io.font.constants.StandardFonts;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

/**
 * Font manager for handling different font styles.
 */
public class FontManager {
    
    private final Map<String, PdfFont> fontCache = new HashMap<>();
    
    private PdfFont regularFont;
    private PdfFont boldFont;
    private PdfFont italicFont;
    private PdfFont boldItalicFont;
    
    public FontManager() throws IOException {
        initializeFonts();
    }
    
    private void initializeFonts() throws IOException {
        // Using Helvetica family for clean, modern look
        regularFont = PdfFontFactory.createFont(StandardFonts.HELVETICA);
        boldFont = PdfFontFactory.createFont(StandardFonts.HELVETICA_BOLD);
        italicFont = PdfFontFactory.createFont(StandardFonts.HELVETICA_OBLIQUE);
        boldItalicFont = PdfFontFactory.createFont(StandardFonts.HELVETICA_BOLDOBLIQUE);
        
        fontCache.put("regular", regularFont);
        fontCache.put("bold", boldFont);
        fontCache.put("italic", italicFont);
        fontCache.put("boldItalic", boldItalicFont);
    }
    
    /**
     * Get the appropriate font based on style flags.
     */
    public PdfFont getFont(boolean bold, boolean italic) {
        if (bold && italic) {
            return boldItalicFont;
        } else if (bold) {
            return boldFont;
        } else if (italic) {
            return italicFont;
        }
        return regularFont;
    }
    
    public PdfFont getRegularFont() {
        return regularFont;
    }
    
    public PdfFont getBoldFont() {
        return boldFont;
    }
    
    public PdfFont getItalicFont() {
        return italicFont;
    }
    
    public PdfFont getBoldItalicFont() {
        return boldItalicFont;
    }
}
