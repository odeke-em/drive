//
//  OperationCacheTests.swift
//
//  Created by Indragie Karunaratne on 8/3/15.
//

import XCTest
@testable import OperationCache

class OperationCacheTests: XCTestCase {
    func testInit() {
        let cache = OperationCache<String, String>()
        for (_, _) in cache.snapshot {
            XCTFail("There should be no values")
        }
    }
    
    func testGetAndSet() {
        let cache = OperationCache<String, String>()
        XCTAssertTrue(cache["foo"] == nil)
        
        let expirable = Expirable(expiryDate: NSDate.distantFuture(), value: "bar")
        cache["foo"] = expirable
        XCTAssertTrue(cache["foo"]! == expirable)
        
        let expirable2 = Expirable(expiryDate: NSDate.distantFuture(), value: "baz")
        cache["foo"] = expirable2
        XCTAssertTrue(cache["foo"]! == expirable2)
    }
    
    func testRemove() {
        let cache = OperationCache<String, String>()
        cache["foo"] = Expirable(expiryDate: NSDate.distantFuture(), value: "bar")
        XCTAssertTrue(cache["foo"] != nil)
        
        cache["foo"] = nil
        XCTAssertTrue(cache["foo"] == nil)
    }
    
    func testExpiration() {
        let cache = OperationCache<String, String>()
        cache["foo"] = Expirable(expiryDate: NSDate.distantPast(), value: "bar")
        XCTAssertTrue(cache["foo"] == nil)
    }
    
    func testSnapshot() {
        let cache = OperationCache<String, String>()
        let expirable = Expirable(expiryDate: NSDate.distantFuture(), value: "bar")
        cache["foo"] = expirable
        
        var generator = cache.snapshot.generate()
        var next = generator.next()
        XCTAssertEqual(next!.0, "foo")
        XCTAssertTrue(next!.1 == expirable)
        
        next = generator.next()
        XCTAssertTrue(next == nil)
    }
}
