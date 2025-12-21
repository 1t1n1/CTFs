package src;



class Main {
    public static void main(String args[]) throws Throwable {
        if (args.length != 1) {
            System.out.println("Usage: java src.Main <input file>");
            System.exit(1);
        }

        String inputFile = args[0];

        var parser = new OrderParser(inputFile);

        for(String order : parser) {
            System.out.println(order);
            System.out.println("");
        }
    }
}
