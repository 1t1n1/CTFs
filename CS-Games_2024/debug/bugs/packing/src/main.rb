
require 'tsort'

if ARGV.length < 1
  puts "Usage: ruby main.rb <input_file>"
  exit 1
end

class Hash
  include TSort
  alias tsort_each_node each_key
  def tsort_each_child(node, &block)
    fetch(node)&.each(&block)
  end
end

require_relative("packing/all")

# read input file
input_file = ARGV[0]
input = Packing::Spec.from_yaml(File.read(input_file))

puts ({ "plan": input.plan! }).to_yaml
