import pytest
from django.contrib.auth.models import User
from django.core.cache import cache
from rest_framework.exceptions import AuthenticationFailed
from ....services import RemoteJWTAuthenticationBackend, ServiceUnavailable
from unittest.mock import Mock, patch
import requests

pytestmark = pytest.mark.django_db

class TestRemoteJWTAuthenticationBackend:
    def setup_method(self):
        self.backend = RemoteJWTAuthenticationBackend()
        cache.clear()

    def test_authenticate_no_token_header(self):
        request = Mock()
        request.headers = {}
        result = self.backend.authenticate(request)
        assert result is None

    def test_authenticate_invalid_header(self):
        request = Mock()
        request.headers = {'Authorization': 'Basic token'}
        result = self.backend.authenticate(request)
        assert result is None

    def test_authenticate_success_with_cache(self, settings):
        settings.REMOTE_AUTH = {'VERIFY_URL': 'http://test/verify', 'TIMEOUT': 5}
        with patch('articles.services.requests.get') as mock_get:
            mock_response = mock_get.return_value
            mock_response.status_code = 200
            mock_response.json.return_value = {
                'data': {
                    'name': 'remoteuser',
                    'code': '1234567890',
                    'image': 'http://example.com/avatar.jpg',
                    'level': 2,
                }
            }
            request = Mock()
            request.headers = {'Authorization': 'Bearer valid.token'}
            user = self.backend.authenticate(request, token='valid.token')
            assert user.username == 'remoteuser'
            assert request.remote_user_payload['username'] == 'remoteuser'
            mock_get.assert_called_once()

    def test_authenticate_timeout(self, settings):
        settings.REMOTE_AUTH = {
            'VERIFY_URL': 'http://test/verify',
            'TIMEOUT': 0.001
        }
        with patch('articles.services.requests.get') as mock_get:
            mock_get.side_effect = requests.exceptions.Timeout()
            request = Mock()
            request.headers = {'Authorization': 'Bearer token'}
            with pytest.raises(ServiceUnavailable):
                self.backend.authenticate(request)

    def test_authenticate_connection_error(self, settings):
        settings.REMOTE_AUTH = {
            'VERIFY_URL': 'http://test/verify',
            'TIMEOUT': 5
        }
        with patch('articles.services.requests.get') as mock_get:
            mock_get.side_effect = requests.exceptions.ConnectionError()
            request = Mock()
            request.headers = {'Authorization': 'Bearer token'}
            with pytest.raises(ServiceUnavailable):
                self.backend.authenticate(request)


    def test_authenticate_invalid_response(self, mock_remote_auth_failure, settings):
        settings.REMOTE_AUTH = {
            'VERIFY_URL': 'http://test/verify',
            'TIMEOUT': 5
        }
        request = Mock()
        request.headers = {'Authorization': 'Bearer invalid.token'}
        with pytest.raises(AuthenticationFailed):
            self.backend.authenticate(request)


    def test_authenticate_missing_user_data(self, settings):
        settings.REMOTE_AUTH = {
            'VERIFY_URL': 'http://test/verify',
            'TIMEOUT': 5
        }
        with patch('articles.services.requests.get') as mock_get:
            mock_response = mock_get.return_value
            mock_response.status_code = 200
            mock_response.json.return_value = {}
            request = Mock()
            request.headers = {'Authorization': 'Bearer token'}
            with pytest.raises(AuthenticationFailed):
                self.backend.authenticate(request)

    def test_get_user(self):
        user = User.objects.create_user(username='test', password='pass')
        result = self.backend.get_user(user.id)
        assert result == user
        result = self.backend.get_user(99999)
        assert result is None