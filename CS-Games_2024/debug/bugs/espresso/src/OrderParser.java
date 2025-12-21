package src;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.io.FileNotFoundException;
import java.util.Iterator;
import java.util.regex.Matcher;
import java.util.regex.Pattern;



class OrderParser implements Iterator<String>, Iterable<String> {
    private BufferedReader reader;
    private int count = 0;

    public OrderParser(String file) throws FileNotFoundException {
        this.reader = new BufferedReader(new FileReader(file));
    }

    @Override
    public boolean hasNext() {
        if(count > 10) {
            return false;
        }

        count++;
        return reader.ready();
    }

    @Override
    public String next() {
        String line = null;
        line = reader.readLine();
        if(line == null) {
            return null;
        }
        if (line.startsWith("Order for:")) {
            return this.parseOrder(line);
        }
        return null;
    }

    @Override
    public void remove() {
        throw new UnsupportedOperationException("Remove operation is not supported");
    }

    @Override
    public Iterator<String> iterator() {
        return this;
    }

    public String parseOrder(String order) {
        String line;
        Pattern p, sp;
        p = Pattern.compile("- (\\d+) (.*) (.*)( .+)?");
        sp = Pattern.compile("(with|and|,)");
        int total = 0;
        do {
            line = reader.readLine();
            if (line == null) break;

            Matcher m = p.matcher(line);
            if (!m.matches()) break;

            int price;
            int quantity = Integer.parseInt(m.group(1));
            String size = m.group(2);
            String drink = m.group(3);
            String options = m.group(4);
            price = drink == "coffee" ? 100 : (drink == "tea" ? 90 : 0);
            Drink d = new Drink(drink, price, size, quantity);
            if (options != null) {
                String[] opts = sp.split(options);
                for (String opt : opts) {
                    if(opt == null || opt.isEmpty()) {
                        continue;
                    }
                    var a = new Addon(opt);
                    d.addons.add(a);
                }
            }
            total += d.calculatePrice();
        } while (line != null);

        order += String.format("\nTotal: $%.2f", total / 100.0);

        return order;
    }
}
