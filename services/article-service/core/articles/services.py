# articles/services.py

import hashlib
import logging
import requests
from django.core.cache import cache
from django.conf import settings
from django.contrib.auth.backends import BaseBackend
from django.contrib.auth.models import User
from rest_framework.exceptions import AuthenticationFailed, APIException

# Set up logger
logger = logging.getLogger(__name__)

# vazife in file ine ke token daryaft shode ro check kone

class ServiceUnavailable(APIException):
    status_code = 503
    default_detail = "Auth service is unavailable."


class RemoteJWTAuthenticationBackend(BaseBackend):
    
    def _get_user_from_auth(self, token):
        logger.info("=" * 50)
        logger.info("_get_user_from_auth: Starting remote token validation")
        logger.info(f"_get_user_from_auth: Token length: {len(token)} characters")
        
        headers = {'Authorization': f'Bearer {token}'}
        logger.debug(f"_get_user_from_auth: Request headers: {headers}")
        
        verify_url = settings.REMOTE_AUTH.get('VERIFY_URL', 'Not configured')
        timeout = settings.REMOTE_AUTH.get('TIMEOUT', 5)
        logger.info(f"_get_user_from_auth: Calling auth service at: {verify_url}")
        logger.info(f"_get_user_from_auth: Timeout: {timeout} seconds")
        
        try:
            logger.debug("_get_user_from_auth: Sending GET request to auth service")
            response = requests.get(
                settings.REMOTE_AUTH['VERIFY_URL'],
                headers=headers,
                timeout=settings.REMOTE_AUTH['TIMEOUT']
            )
            logger.info(f"_get_user_from_auth: Response status code: {response.status_code}")
            logger.debug(f"_get_user_from_auth: Response text: {response.text[:200] if response.text else 'Empty response'}")
            
            response.raise_for_status()
            logger.info("_get_user_from_auth: Response status OK")
            
        except requests.exceptions.Timeout as e:
            logger.error(f"_get_user_from_auth: ❌ Timeout error: {str(e)}")
            raise ServiceUnavailable("Auth service timeout.")
        except requests.exceptions.ConnectionError as e:
            logger.error(f"_get_user_from_auth: ❌ Connection error: {str(e)}")
            raise ServiceUnavailable("Auth service unreachable.")
        except requests.exceptions.RequestException as e:
            logger.error(f"_get_user_from_auth: ❌ Request exception: {str(e)}")
            if hasattr(e, 'response') and e.response is not None:
                logger.error(f"_get_user_from_auth: Response status: {e.response.status_code}")
                logger.error(f"_get_user_from_auth: Response body: {e.response.text[:200]}")
            raise AuthenticationFailed(f"Token verification failed: {str(e)}")

        logger.info("_get_user_from_auth: Parsing JSON response")
        result = response.json()
        logger.debug(f"_get_user_from_auth: Full response JSON: {result}")
        
        # Extract user data from response
        user_data = result.get('data')
        if not user_data:
            logger.error("_get_user_from_auth: ❌ Missing 'data' field in response")
            logger.error(f"_get_user_from_auth: Response keys: {result.keys() if isinstance(result, dict) else 'Not a dict'}")
            raise AuthenticationFailed("Invalid token payload: missing 'data' field.")
        
        logger.info(f"_get_user_from_auth: User data extracted successfully")
        logger.info(f"_get_user_from_auth: User fields - name: {user_data.get('name', 'N/A')}, code: {user_data.get('code', 'N/A')}")
        
        mapped = {
            'username': user_data.get('name', ''),
            'citizenId': user_data.get('code', ''),
            'avatar': user_data.get('image', ''),
            'slug_level': user_data.get('level', 0),      # موقتی
            'image_level': '',   # فعلاً خالی
        }
        
        logger.info(f"_get_user_from_auth: Mapped user data: {mapped}")
        logger.info("_get_user_from_auth: Remote validation completed successfully")
        logger.info("=" * 50)
        return mapped

    def authenticate(self, request, token=None, **kwargs):
        logger.info("=" * 60)
        logger.info("🚀 RemoteJWTAuthenticationBackend.authenticate: START")
        logger.info(f"Request path: {request.path if hasattr(request, 'path') else 'Unknown'}")
        logger.info(f"Request method: {request.method if hasattr(request, 'method') else 'Unknown'}")
        
        # gereftan token az header
        if token is None:
            logger.debug("No token parameter provided, checking Authorization header")
            auth_header = request.headers.get('Authorization', '')
            
            # Log header safely (mask if too long)
            if len(auth_header) > 100:
                logger.info(f"Authorization header (first 100 chars): '{auth_header[:100]}...'")
            else:
                logger.info(f"Authorization header: '{auth_header}'")
            
            if not auth_header:
                logger.warning("⚠️ No Authorization header found in request")
                logger.info("Authentication failed: Missing Authorization header")
                logger.info("=" * 60)
                return None
                
            if not auth_header.startswith('Bearer '):
                logger.warning(f"⚠️ Authorization header doesn't start with 'Bearer '")
                logger.warning(f"Header starts with: '{auth_header[:15]}...'")
                logger.info("Authentication failed: Invalid auth header format")
                logger.info("=" * 60)
                return None
                
            token = auth_header.split(' ')[1]
            logger.info(f"✅ Token extracted from header (length: {len(token)} characters)")
            logger.debug(f"Token preview: {token[:20]}...{token[-10:] if len(token) > 30 else token}")
        else:
            logger.info(f"Token provided as parameter (length: {len(token)} characters)")
            logger.debug(f"Token preview: {token[:20]}...{token[-10:] if len(token) > 30 else token}")

        # cach kardan baraye kahesh request be auth service
        cache_key = f"jwt_payload_{hashlib.md5(token.encode()).hexdigest()}"
        logger.debug(f"Cache key generated: {cache_key}")
        
        logger.info("Checking cache for existing payload...")
        cached_payload = cache.get(cache_key)
        
        if cached_payload:
            logger.info("✅✅✅ Cache HIT! Using cached user payload (avoiding remote call)")
            logger.info(f"Cached payload username: {cached_payload.get('username', 'Unknown')}")
            logger.debug(f"Full cached payload: {cached_payload}")
            
            request.remote_user_payload = cached_payload
            logger.debug("Stored cached payload in request.remote_user_payload")
            
            user, created = User.objects.get_or_create(username=cached_payload['username'])
            if created:
                logger.info(f"📝 Created NEW local user from cache: {cached_payload['username']}")
            else:
                logger.info(f"✅ Retrieved EXISTING local user from cache: {cached_payload['username']}")
            
            logger.info("Authentication successful (CACHED)")
            logger.info("=" * 60)
            return user
        else:
            logger.info("❌ Cache MISS - No cached payload found")
            logger.info("Will call remote auth service for validation")

        # gereftan data user az auth service
        logger.info("Calling _get_user_from_auth to validate token remotely...")
        try:
            user_payload = self._get_user_from_auth(token)
            logger.info("✅ Remote authentication successful!")
            logger.info(f"Remote user payload: {user_payload}")
        except Exception as e:
            logger.error(f"❌❌❌ Remote authentication FAILED: {str(e)}", exc_info=True)
            logger.info("Authentication failed: Invalid token or service error")
            logger.info("=" * 60)
            raise  # Re-raise the exception to let DRF handle it

        # zakhire dar cach (5 daghighe)
        logger.info(f"Storing payload in cache for 300 seconds (5 minutes)")
        cache.set(cache_key, user_payload, timeout=300)
        logger.debug(f"Cached payload: {user_payload}")

        # zakhire dar request baraye estefdae dar view
        request.remote_user_payload = user_payload
        logger.debug("Stored user payload in request.remote_user_payload")

        # Get or create local user
        logger.info(f"Looking up local user with username: {user_payload['username']}")
        user, created = User.objects.get_or_create(username=user_payload['username'])
        
        if created:
            logger.info(f"📝 Created NEW local user: {user_payload['username']} (ID: {user.id})")
        else:
            logger.info(f"✅ Retrieved EXISTING local user: {user_payload['username']} (ID: {user.id})")
        
        logger.info("🎉 Authentication completed SUCCESSFULLY!")
        logger.info("=" * 60)
        return user

    def get_user(self, user_id):
        logger.debug(f"get_user called with user_id: {user_id}")
        try:
            user = User.objects.get(pk=user_id)
            logger.debug(f"User found: {user.username} (ID: {user_id})")
            return user
        except User.DoesNotExist:
            logger.warning(f"User with ID {user_id} does not exist in local database")
            return None