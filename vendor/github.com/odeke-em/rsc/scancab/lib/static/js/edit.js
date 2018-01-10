var param = function(obj)
    {
      var query = '';
      var name, value, fullSubName, subValue, innerObj, i;
      
      for(name in obj)
      {
        value = obj[name];
        
        if(value instanceof Array)
        {
          for(i=0; i<value.length; ++i)
          {
            subValue = value[i];
            fullSubName = name + '[' + i + ']';
            innerObj = {};
            innerObj[fullSubName] = subValue;
            query += param(innerObj) + '&';
          }
        }
        else if(value instanceof Object)
        {
          for(subName in value)
          {
            subValue = value[subName];
            fullSubName = name + '[' + subName + ']';
            innerObj = {};
            innerObj[fullSubName] = subValue;
            query += param(innerObj) + '&';
          }
        }
        else if(value !== undefined && value !== null)
        {
          query += encodeURIComponent(name) + '=' + encodeURIComponent(value) + '&';
        }
      }
      
      return query.length ? query.substr(0, query.length - 1) : query;
    };

function SearchBoxCtrl($scope, $location) {
  $scope.query = "";
  var q = $location.search()["q"];
  if (q !== undefined) {
    $scope.query = q;
  }
  $scope.search = function() {
    $location.search("q", $scope.query);
  }
}

function dopdf($scope) {
  $scope.showingpdf = false;
  $scope.showpdf = function() {
    var url = '/scan/doc/' + $scope.docid + '/pdf';
    if(new PDFObject({url: '/scan/doc/' + $scope.docid + '/pdf'}).embed('pdf')) {
      $scope.showingpdf = true;
      $('html, body').animate({scrollTop:$('#pdftop').offset().top}, 'slow');
    } else {
      window.open(url, '_blank');
    }
  }

  $scope.hidepdf = function() {
    $scope.showingpdf = false;
  }
}

function EditCtrl($scope, $http, $routeParams, $location) {
  $scope.docid = $routeParams.docid;
  $http.get('/scan/doc/' + $scope.docid).success(function(data) {
    var zeroTime = "0001-01-01T00:00:00Z";
    if (data.Time == zeroTime) { data.Time = ""; }
    if (data.Due == zeroTime) { data.Due = ""; }
    $scope.doc = data;
  });
  $scope.submit = function() {
    $http({
      method: 'POST',
      url: '/scan/doc/'+$scope.docid+'/edit',
      data: param($scope.doc),
      headers: {'Content-Type': 'application/x-www-form-urlencoded'}
    }).success(function() {
      if ($scope.incoming) {
        $location.path('/incoming');
      } else {
        $location.path('/doc/' + $scope.docid);
      }
    }).error(function() {
      $scope.message = "FAILED TO SAVE";
    })
  }
  
  dopdf($scope);
}

function EditIncomingCtrl($scope, $http, $routeParams, $location) {
  $scope.incoming = true;
  $scope.incoming_count = "?";
  $http.get('/scan/incoming-count').success(function(data) {
    $scope.incoming_count = data;
  })
  EditCtrl($scope, $http, $routeParams, $location);
}

function ShowCtrl($scope, $http, $routeParams) {
  $scope.docid = $routeParams.docid;
  $http.get('/scan/doc/' + $scope.docid).success(function(data) {
    var zeroTime = "0001-01-01T00:00:00Z";
    if (data.Time == zeroTime) { data.Time = ""; }
    if (data.Due == zeroTime) { data.Due = ""; }
    $scope.doc = data;
  });
  dopdf($scope);
}

function ListCtrl($scope, $http, $location) {
  $scope.query = "";
  var q = $location.search()["q"];
  if (q !== undefined) {
    $scope.query = q;
  }
  $scope.start = 0
  $scope.n = $location.search()["n"];
  if ($scope.n === undefined) {
    $scope.n = 10;
  }
  $scope.load = function() {
    $scope.more = false;
    $http.get('/scan/list?Start=' + $scope.start + '&N=' + ($scope.n+1) + '&Q=' + $scope.query).success(function(data) {
      if (data.length > $scope.n) {
        data = data.slice(0, $scope.n);
        $scope.more = true;
      }
      var zeroTime = "0001-01-01T00:00:00Z";
      for (var i = 0; i < data.length; i++) {
        var doc = data[i];
        if (doc.Time == zeroTime) { doc.Time = ""; }
        if (doc.Due == zeroTime) { doc.Due = ""; }
      }
      $scope.docs = data;
    });
  }
  $scope.load();

  $scope.next = function() { $scope.start += 10; $scope.load(); }
  $scope.prev = function() { if($scope.start > 0) $scope.start -= 10; $scope.load(); }
}
