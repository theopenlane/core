# Pooling control

Be certain to review the [pond](https://github.com/alitto/pond) library which is what this package is based on. There are a number of controls and options you can provide to the pool to control it's behavior, resizing, etc - [ref](https://github.com/alitto/pond?tab=readme-ov-file#pool-configuration-options). You can also find / see some of the pond benchmarks for performance here: https://github.com/alitto/pond-benchmarks.

## Resizing strategies

Configures the strategy used to resize the pool when backpressure is detected. You can create a custom strategy by implementing the pond.ResizingStrategy interface or choose one of the 3 presets:

- Eager: maximizes responsiveness at the expense of higher resource usage, which can reduce throughput under certain conditions. This strategy is meant for worker pools that will operate at a small percentage of their capacity most of the time and may occasionally receive bursts of tasks. This is the default strategy.
- Balanced: tries to find a balance between responsiveness and throughput. It's suitable for general purpose worker pools or those that will operate close to 50% of their capacity most of the time.
- Lazy: maximizes throughput at the expense of responsiveness. This strategy is meant for worker pools that will operate close to their max. capacity most of the time.

```go
eagerPool := soiree.NewPondPool(10, 1000, pond.Strategy(pond.Eager()))
balancedPool := soiree.New(10, 1000, pond.Strategy(pond.Balanced()))
lazyPool := soiree.New(10, 1000, pond.Strategy(pond.Lazy()))
```