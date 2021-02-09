//
//  OperationCache.swift
//
//  Created by Indragie Karunaratne on 8/3/15.
//

import Foundation

public struct Expirable<ValueType> {
    public let expiryDate: NSDate
    public let value: ValueType
    
    public var expired: Bool {
        return NSDate().compare(expiryDate) == .OrderedDescending
    }
}

public func ==<ValueType: Equatable>(lhs: Expirable<ValueType>, rhs: Expirable<ValueType>) -> Bool {
    return lhs.expiryDate == rhs.expiryDate && lhs.value == rhs.value
}

/// Swift port of https://github.com/odeke-em/cache/blob/master/cache.go
///
/// Lockless & thread safe with support for multiple concurrent readers
/// and a single writer.
public class OperationCache<KeyType: Hashable, ValueType> {
    public typealias ExpirableType = Expirable<ValueType>
    private let queue = dispatch_queue_create("com.indragie.OperationCache", DISPATCH_QUEUE_CONCURRENT)
    private var storage = [KeyType: ExpirableType]()
    
    public var snapshot: Zip2Sequence<[KeyType], [ExpirableType]> {
        var sequence: Zip2Sequence<[KeyType], [ExpirableType]>? = nil
        dispatch_sync(queue) {
            sequence = zip(self.storage.keys.array, self.storage.values.array)
        }
        return sequence!
    }
    
    public subscript(key: KeyType) -> Expirable<ValueType>? {
        get {
            var value: ExpirableType?
            dispatch_sync(queue) { value = self.storage[key] }
            if let value = value where value.expired {
                self[key] = nil
                return nil
            }
            return value
        }
        set {
            dispatch_barrier_async(queue) { self.storage[key] = newValue }
        }
    }
}
