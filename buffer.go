package logaggregator

// type RingBuffer struct {
//     data     []*LogEntry
//     size     int
//     writePos int  // where the NEXT write goes
//     count    int  // how many entries currently stored (capped at size)
// }

// func (r *RingBuffer) Insert(entry LogEntry) {
//     r.data[r.writePos] = &entry
//     r.writePos = (r.writePos + 1) % r.size

//     if r.count < r.size {
//         r.count++
//     }
//     // if count == size, we just overwrote the oldest — count stays the same
// }

// func (r *RingBuffer) Query() []*LogEntry {
//     result := make([]*LogEntry, 0, r.count)

//     // oldest entry's index — this is the one formula
//     start := (r.writePos - r.count + r.size) % r.size

//     for i := 0; i < r.count; i++ {
//         idx := (start + i) % r.size
//         result = append(result, r.data[idx])
//     }
//     return result
// }