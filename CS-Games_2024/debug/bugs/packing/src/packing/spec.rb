require 'yaml'

module Packing
  class Spec
    attr_reader :workers, :jobs, :strategy

    def initialize(workers, jobs, strategy)
      @workers = workers
      @jobs = jobs
      @strategy = Strategy.find(strategy).new(self)
    end

    def self.from_yaml(document)
      yaml = YAML.safe_load(document)

      workers = yaml["spec"]["workers"].map do |worker|
        Worker.from(worker)
      end

      jobs = yaml["spec"]["jobs"].map do |job|
        Job.from(job)
      end

      Spec.new(workers, jobs, yaml["spec"]["strategy"])
    end

    def to_s
      YAML.dump(self)
    end

    def plan!
      begin
        @strategy.plan!
      rescue e
        { "error": e.message }
      end
    end
  end

end
