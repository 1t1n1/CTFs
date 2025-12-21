module Packing
  autoload :Strategies, 'packing/strategies/all'

  class Strategy
    attr_reader :spec

    def self.find(name)
      Packing::Strategies.const_get(name)
    end

    def initialize(spec)
      @spec = spec
    end

    def plan!
      jobs_by_id = @spec.jobs.map { |job| [job.id, job] }.to_h
      jobs = job_order.map do |job|
        jobs_by_id[job]
      end

      jobs.map do |job, i|
        schedule(job)

        {
          priority: i + 1,
          job_id: job.id,
          worker_id: job.worker
        }
      end
    end

    def job_order
      @dependencies = {}

      @spec.jobs.each do |job|
        @dependencies[job.id] = job.depends_on
      end

      @dependencies.tsort
    rescue TSort::Cyclic
      raise 'Cyclic dependencies detected'
    end
  end
end
