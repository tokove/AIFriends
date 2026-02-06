from django.contrib.auth.models import User
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework_simplejwt.tokens import RefreshToken

from web.models.user import UserProfile


class RegisterView(APIView):
    def post(self, request):
        try:
            username = request.data['username'].strip()
            password = request.data['password'].strip()
            if not username or not password:
                return Response({
                    "result": "用户名和密码不能为空"
                })
            if User.objects.filter(username=username).exists():
                return Response({
                    "result": "用户名已存在"
                })
            user = User.objects.create_user(username=username, password=password) # 自动创建用户
            user_profile = UserProfile.objects.create(user=user)
            refresh = RefreshToken.for_user(user)
            response = Response({
                'result': 'success',
                'access': str(refresh.access_token),
                'user_id': user.id,
                'username': user.username,
                'photo': user_profile.photo.url,  # 必须加url，否则返回文件名称
                'profile': user_profile.profile,
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
        except:
            import traceback
            print(traceback.format_exc())
            return Response({
                "result": "系统繁忙，请稍后再试"
            })
