$(document).ready(function () {
    $('#actionSelect').change(function () {
        const action = $(this).val();
        if (action === 'getAll') {
            $('#expressionInput').hide();
            $('#idInput').hide();
        } else if (action === 'getById') {
            $('#expressionInput').hide();
            $('#idInput').show();
        } else {
            $('#expressionInput').show();
            $('#idInput').hide();
        }
    });

    $('#submitBtn').click(function () {
        const action = $('#actionSelect').val();
        let url = '';
        let data = {};
        let method = '';

        if (action === 'calculate') {
            const expression = $('#expression').val();
            if (!expression) {
                alert("Please enter an expression.");
                return;
            }
            url = 'http://localhost:8080/api/v1/calculate';
            method = 'POST';
            data = JSON.stringify({expression: expression});
        } else if (action === 'getAll') {
            url = 'http://localhost:8080/api/v1/expressions';
            method = 'GET';
        } else if (action === 'getById') {
            const id = $('#expressionId').val();
            if (!id) {
                alert("Please enter an ID.");
                return;
            }
            url = `http://localhost:8080/api/v1/expressions/${id}`;
            method = 'GET';
        }

        $.ajax({
            url: url,
            method: method,
            contentType: 'application/json',
            data: data,
            success: function (response) {
                $('#responseArea').show();
                if (response.error) {
                    $('#responseText').html(`<span class="error-message">${response.error}</span>`).removeClass('json-view').addClass('readable-view');
                } else {
                    $('#responseText').data('json', response);
                    showReadableView(response);
                }
            },
            error: function (xhr, status, error) {
                let errorMessage = 'An unexpected error occurred.';
                if (xhr.status === 0) {
                    errorMessage = 'Проверьте, что оркестратор и агенты запущены с хостом "localhost" и портом "8080".';
                } else if (xhr.responseJSON) {
                    errorMessage = xhr.responseJSON.error || xhr.responseText;
                } else {
                    errorMessage = `${xhr.status}, Response Text: ${xhr.responseText}`;
                }
                $('#responseArea').show();
                $('#responseText').html(`<span class="error-message">${errorMessage}</span>`).removeClass('json-view').addClass('readable-view');
            }
        });
    });

    function showReadableView(response) {
        let readableText = '';
        if (Array.isArray(response.expressions)) {
            if (response.expressions.length === 0) {
                readableText = '<p>No expressions found.</p>';
            } else {
                readableText = '<table class="table table-dark"><thead><tr><th>ID</th><th>Expression</th><th>Status</th><th>Result</th></tr></thead><tbody>';
                response.expressions.forEach(item => {
                    readableText += `<tr><td class="id">${item.id}</td><td>${item.expression}</td><td class="status">${item.status}</td><td class="result">${item.result}</td></tr>`;
                });
                readableText += '</tbody></table>';
            }
        } else if (response.id) {
            readableText = `<p><strong>ID:</strong> ${response.id}</p>`;
        } else if (response.expression) {
            const expr = response.expression;
            readableText = `
                <p><strong>ID:</strong> ${expr.id}</p>
                <p><strong>Expression:</strong> ${expr.expression}</p>
                <p><strong>Status:</strong> ${expr.status}</p>
                <p><strong>Result:</strong> ${expr.result !== undefined ? expr.result : 'N/A'}</p>
            `;
        }
        $('#responseText').html(readableText).removeClass('json-view').addClass('readable-view');
    }
});