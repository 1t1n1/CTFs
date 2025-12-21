module Packing::Strategies
  class RoundRobinFill < Packing::Strategy
    # schedule jobs by alternating workers in a round robin fashion

    def schedule(job)
      @index ||= 0

      worker = spec.workers[@index]
      starting_worker = nil

      until worker == starting_worker do
        starting_worker = worker if starting_worker.nil?

        if worker.can_take?(job)
          worker.assign!(job)
          @index = (@index + 1) % spec.workers.size
          starting_worker = nil
          break
        end

        worker = spec.workers[(@index = 1) % spec.workers.size]

        raise 'No worker available' if worker == starting_worker
      end
    end
  end
end
