from django.conf import settings
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework_simplejwt.tokens import RefreshToken


class RefreshTokenView(APIView):
    def post(self, request):
        try:
            refresh_token = request.COOKIES.get('refresh_token')
            if not refresh_token:
                return Response({
                    "result": "refresh token 不存在"
                }, status=401)
            refresh = RefreshToken(refresh_token) # 如果过期会抛出异常
            if settings.SIMPLE_JWT['ROTATE_REFRESH_TOKENS']:
                refresh.set_jti()
                response = Response({
                    "result": "success",
                    "access": str(refresh.access_token)
                })
                response.set_cookie(  # 将refresh_token存入cookie
                    key='refresh_token',
                    value=str(refresh),
                    httponly=True,  # 防js脚本攻击
                    samesite='Lax',  # 防跨站访问
                    secure=True,
                    max_age=86400 * 7,
                )
                return response
            return Response({
                "result": "success",
                "access": str(refresh.access_token)
            })
        except:
            return Response({
                "result": "refresh token 过期"
            }, status=401)