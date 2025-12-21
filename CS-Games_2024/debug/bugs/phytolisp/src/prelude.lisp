;; map :: (lambda (a b)) -> (list a) -> (list b)
(set map
  (lambda (f xs)
    (if :xs (cons (:f (head :xs)) (:map :f (tail :xs))) ())))

;; filter :: (lambda (a boolean)) -> (list a)
(set filter
  (lambda (f xs)
    (if :xs
        (if (:f (head :xs)
            (cons (head :xs)
                  (filter :f (tail :xs)))
            (filter :f (tail :xs)))
        ())))

;; reduce :: (lambda (a b)) -> a -> (list b) -> a
(set reduce
    (lambda (f x xs)
        (if :xs
            (:reduce :f (:f :x (head :xs) (tail :xs))
          :x)))))
