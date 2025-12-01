package com.amazonreceipt;

import com.amazonreceipt.generator.PdfReceiptGenerator;
import com.amazonreceipt.model.ReceiptData;
import com.amazonreceipt.util.TimerUtil;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;

import java.io.FileReader;
import java.io.IOException;
import java.io.Reader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * Amazon Receipt PDF Generator
 * 
 * A beautiful PDF receipt generator using iText 8.
 * Reads JSON receipt data and generates professional PDF documents.
 * 
 * @author Amazon Receipt Generator Team
 * @version 1.0.0
 */
public class AmazonReceiptGenerator {

    private static final String BANNER = 
        "\n" +
        "╔═══════════════════════════════════════════════════════════════════════════╗\n" +
        "║                                                                           ║\n" +
        "║     █████╗ ███╗   ███╗ █████╗ ███████╗ ██████╗ ███╗   ██╗                ║\n" +
        "║    ██╔══██╗████╗ ████║██╔══██╗╚══███╔╝██╔═══██╗████╗  ██║                ║\n" +
        "║    ███████║██╔████╔██║███████║  ███╔╝ ██║   ██║██╔██╗ ██║                ║\n" +
        "║    ██╔══██║██║╚██╔╝██║██╔══██║ ███╔╝  ██║   ██║██║╚██╗██║                ║\n" +
        "║    ██║  ██║██║ ╚═╝ ██║██║  ██║███████╗╚██████╔╝██║ ╚████║                ║\n" +
        "║    ╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝                ║\n" +
        "║                                                                           ║\n" +
        "║              📄 RECEIPT PDF GENERATOR - iText 8 Edition                   ║\n" +
        "║                                                                           ║\n" +
        "╚═══════════════════════════════════════════════════════════════════════════╝\n";

    private static final String DEFAULT_INPUT = "../../sampledata/amazon/amazon_receipt.json";
    private static final String DEFAULT_OUTPUT = "amazon_receipt_output.pdf";

    public static void main(String[] args) {
        System.out.println(BANNER);
        
        // Parse command line arguments
        String inputPath = args.length > 0 ? args[0] : DEFAULT_INPUT;
        String outputPath = args.length > 1 ? args[1] : DEFAULT_OUTPUT;

        // Start overall timer
        TimerUtil overallTimer = new TimerUtil("PDF Generation (Total)");
        overallTimer.start();

        try {
            // Timer for JSON parsing
            TimerUtil jsonTimer = new TimerUtil("JSON Parsing");
            jsonTimer.start();
            
            // Read and parse JSON
            ReceiptData receiptData = parseJsonFile(inputPath);
            
            jsonTimer.stop();

            if (receiptData == null) {
                printError("Failed to parse JSON file: " + inputPath);
                return;
            }

            printInfo("Successfully parsed receipt data");
            printInfo("Title: " + (receiptData.getTitle() != null ? receiptData.getTitle().getText() : "N/A"));
            printInfo("Tables: " + (receiptData.getTables() != null ? receiptData.getTables().size() : 0));

            // Timer for PDF generation
            TimerUtil pdfTimer = new TimerUtil("PDF Document Generation");
            pdfTimer.start();

            // Generate PDF
            PdfReceiptGenerator generator = new PdfReceiptGenerator();
            generator.generate(receiptData, outputPath);

            pdfTimer.stop();

            // Verify output
            Path outputFile = Paths.get(outputPath);
            if (Files.exists(outputFile)) {
                long fileSize = Files.size(outputFile);
                printSuccess("PDF generated successfully!");
                printInfo("Output file: " + outputFile.toAbsolutePath());
                printInfo("File size: " + formatFileSize(fileSize));
            }

        } catch (Exception e) {
            printError("Error during PDF generation: " + e.getMessage());
            e.printStackTrace();
        } finally {
            // Stop overall timer
            overallTimer.stop();
        }

        printFooter();
    }

    /**
     * Parse JSON file into ReceiptData object.
     */
    private static ReceiptData parseJsonFile(String filePath) throws IOException {
        Gson gson = new GsonBuilder()
            .setPrettyPrinting()
            .create();

        Path path = Paths.get(filePath);
        
        // Try to resolve the path
        if (!Files.exists(path)) {
            // Try relative to current directory
            path = Paths.get(System.getProperty("user.dir"), filePath);
        }

        if (!Files.exists(path)) {
            printError("Input file not found: " + filePath);
            printInfo("Tried paths:");
            printInfo("  - " + Paths.get(filePath).toAbsolutePath());
            printInfo("  - " + path.toAbsolutePath());
            return null;
        }

        printInfo("Reading JSON from: " + path.toAbsolutePath());

        try (Reader reader = new FileReader(path.toFile())) {
            return gson.fromJson(reader, ReceiptData.class);
        }
    }

    /**
     * Format file size to human readable string.
     */
    private static String formatFileSize(long bytes) {
        if (bytes < 1024) return bytes + " B";
        if (bytes < 1024 * 1024) return String.format("%.2f KB", bytes / 1024.0);
        return String.format("%.2f MB", bytes / (1024.0 * 1024.0));
    }

    /**
     * Print info message.
     */
    private static void printInfo(String message) {
        System.out.println("  ℹ️  " + message);
    }

    /**
     * Print success message.
     */
    private static void printSuccess(String message) {
        System.out.println("\n  ✅ " + message);
    }

    /**
     * Print error message.
     */
    private static void printError(String message) {
        System.out.println("\n  ❌ ERROR: " + message);
    }

    /**
     * Print footer.
     */
    private static void printFooter() {
        System.out.println(
        "\n" +
        "╔═══════════════════════════════════════════════════════════════════════════╗\n" +
        "║                     🎉 Process Complete!                                  ║\n" +
        "║                                                                           ║\n" +
        "║  Built with iText 8 - The leading PDF library for Java                    ║\n" +
        "║  https://itextpdf.com                                                     ║\n" +
        "╚═══════════════════════════════════════════════════════════════════════════╝\n"
        );
    }
}
