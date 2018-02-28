
var showCarRemovalBox = function(){
    $('#confirmDeleteBox').show();
}

var confirmCarRemoval = function(carId){
    $.ajax({
        url: '/cars/delete',
        type: 'POST',
        data: {car_id: carId}
    });
    $('#confirmDeleteBox').hide();
}

var cancelCarRemoval = function(){
    $('#confirmDeleteBox').hide();
}
