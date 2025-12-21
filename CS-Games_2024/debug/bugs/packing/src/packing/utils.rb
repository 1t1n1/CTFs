module Packing
  module WithAttributes
    def self.included(base)
      base.extend(ClassMethods)
    end

    module ClassMethods
      def attributes(*attrs)
        @attributes ||= Set.new

        attrs.each do |attr|
          @attributes << attr
          attr_reader attr
        end
      end

      def from(hash)
        kwargs = @attributes.map do |attr|
          [attr, hash[attr]]
        end.to_h

        new(**kwargs)
      end
    end

    def initialize(**kwargs)
      kwargs.each do |key, value|
        instance_variable_set("@#{key}", value)
      end
    end

    def to_s
      YAML.dump(self)
    end
  end
end
