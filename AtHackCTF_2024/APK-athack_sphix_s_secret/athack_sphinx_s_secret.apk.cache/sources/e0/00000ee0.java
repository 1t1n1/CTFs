package com.athack.ctf;

import android.content.SharedPreferences;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.Toast;
import androidx.appcompat.app.AppCompatActivity;
import androidx.constraintlayout.widget.ConstraintLayout;
import com.google.android.material.textfield.TextInputLayout;
import kotlin.Metadata;
import kotlin.jvm.internal.Intrinsics;

/* compiled from: MainActivity.kt */
@Metadata(d1 = {"\u0000,\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\u0018\u00002\u00020\u0001B\u0005¢\u0006\u0002\u0010\u0002J\u0012\u0010\r\u001a\u00020\u000e2\b\u0010\u000f\u001a\u0004\u0018\u00010\u0010H\u0014R\u001a\u0010\u0003\u001a\u00020\u0004X\u0086.¢\u0006\u000e\n\u0000\u001a\u0004\b\u0005\u0010\u0006\"\u0004\b\u0007\u0010\bR\u000e\u0010\t\u001a\u00020\nX\u0082.¢\u0006\u0002\n\u0000R\u000e\u0010\u000b\u001a\u00020\fX\u0082.¢\u0006\u0002\n\u0000¨\u0006\u0011"}, d2 = {"Lcom/athack/ctf/MainActivity;", "Landroidx/appcompat/app/AppCompatActivity;", "()V", "dbHelper", "Lcom/athack/ctf/DatabaseHelper;", "getDbHelper", "()Lcom/athack/ctf/DatabaseHelper;", "setDbHelper", "(Lcom/athack/ctf/DatabaseHelper;)V", "loginButton", "Landroid/widget/Button;", "passwordTextInputLayout", "Lcom/google/android/material/textfield/TextInputLayout;", "onCreate", "", "savedInstanceState", "Landroid/os/Bundle;", "app_release"}, k = 1, mv = {1, 8, 0}, xi = ConstraintLayout.LayoutParams.Table.LAYOUT_CONSTRAINT_VERTICAL_CHAINSTYLE)
/* loaded from: classes.dex */
public final class MainActivity extends AppCompatActivity {
    public DatabaseHelper dbHelper;
    private Button loginButton;
    private TextInputLayout passwordTextInputLayout;

    public final DatabaseHelper getDbHelper() {
        DatabaseHelper databaseHelper = this.dbHelper;
        if (databaseHelper != null) {
            return databaseHelper;
        }
        Intrinsics.throwUninitializedPropertyAccessException("dbHelper");
        return null;
    }

    public final void setDbHelper(DatabaseHelper databaseHelper) {
        Intrinsics.checkNotNullParameter(databaseHelper, "<set-?>");
        this.dbHelper = databaseHelper;
    }

    /* JADX INFO: Access modifiers changed from: protected */
    @Override // androidx.fragment.app.FragmentActivity, androidx.activity.ComponentActivity, androidx.core.app.ComponentActivity, android.app.Activity
    public void onCreate(Bundle bundle) {
        super.onCreate(bundle);
        setContentView(R.layout.activity_main);
        SharedPreferences.Editor edit = getSharedPreferences("athack", 0).edit();
        edit.putString("key", "QVRIQUNLQ1RGe3FZZlJqREk5c2hjTnZ5ZGI0YUFIZ280WX0=");
        edit.putString("value", "athackctf@gmail.com");
        setDbHelper(new DatabaseHelper(this));
        getDbHelper().getWritableDatabase();
        View findViewById = findViewById(R.id.loginButton);
        Intrinsics.checkNotNullExpressionValue(findViewById, "findViewById(R.id.loginButton)");
        Button button = (Button) findViewById;
        this.loginButton = button;
        TextInputLayout textInputLayout = null;
        if (button == null) {
            Intrinsics.throwUninitializedPropertyAccessException("loginButton");
            button = null;
        }
        button.setOnClickListener(new View.OnClickListener() { // from class: com.athack.ctf.MainActivity$$ExternalSyntheticLambda0
            @Override // android.view.View.OnClickListener
            public final void onClick(View view) {
                MainActivity.onCreate$lambda$0(MainActivity.this, view);
            }
        });
        View findViewById2 = findViewById(R.id.passwordTextInputLayout);
        Intrinsics.checkNotNullExpressionValue(findViewById2, "findViewById(R.id.passwordTextInputLayout)");
        TextInputLayout textInputLayout2 = (TextInputLayout) findViewById2;
        this.passwordTextInputLayout = textInputLayout2;
        if (textInputLayout2 == null) {
            Intrinsics.throwUninitializedPropertyAccessException("passwordTextInputLayout");
        } else {
            textInputLayout = textInputLayout2;
        }
        textInputLayout.setEndIconOnClickListener(new View.OnClickListener() { // from class: com.athack.ctf.MainActivity$$ExternalSyntheticLambda1
            @Override // android.view.View.OnClickListener
            public final void onClick(View view) {
                MainActivity.onCreate$lambda$1(MainActivity.this, view);
            }
        });
    }

    /* JADX INFO: Access modifiers changed from: private */
    public static final void onCreate$lambda$0(MainActivity this$0, View view) {
        Intrinsics.checkNotNullParameter(this$0, "this$0");
        Toast.makeText(this$0, "You didn't get it yet try again...", 1).show();
    }

    /* JADX INFO: Access modifiers changed from: private */
    public static final void onCreate$lambda$1(MainActivity this$0, View view) {
        Intrinsics.checkNotNullParameter(this$0, "this$0");
        Toast.makeText(this$0, "Hint password", 1).show();
    }
}