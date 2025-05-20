/*
Copyright 2024 xiloss.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/crypto/bcrypt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nats-io/jwt"
	"github.com/nats-io/nats.go"
	"github.com/xiloss/kubernats/api/auth/v1alpha1"
	authv1alpha1 "github.com/xiloss/kubernats/api/auth/v1alpha1"
)

// NATSUserAccountReconciler reconciles a NATSUserAccount object
type NATSUserAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type User struct {
  Pass        string
  Account     string
  Permissions jwt.Permissions
}

// +kubebuilder:rbac:groups=auth.kubernats.ai,resources=natsuseraccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auth.kubernats.ai,resources=natsuseraccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=auth.kubernats.ai,resources=natsuseraccounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NATSUserAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *NATSUserAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
  nc, err := nats.Connect("nats://nats:4222")
  if err != nil {
    return ctrl.Result{}, err
  }

  js, err := nc.JetStream()
  if err != nil {
    return ctrl.Result{}, err
  }

  kv, err := js.KeyValue("auth-info")
  if err == nats.ErrBucketNotFound {
    kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
      Bucket: "auth-info",
      History: 1,
    })
  }

  if err != nil {
    return ctrl.Result{}, err
  }
  
  const finalizer = "auth.kubernats.io/finalizer"

  instance := &v1alpha1.NATSUserAccount{}
  if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
  }

  if instance.DeletionTimestamp.IsZero() {
    if !controllerutil.ContainsFinalizer(instance, finalizer) {
      controllerutil.AddFinalizer(instance, finalizer)
      r.Update(ctx, instance)
    }
  } else {
    rawData, err := kv.Get("credentials")
    if err != nil {
      return ctrl.Result{}, err
    }
    var users map[string]any
    if err := json.Unmarshal(rawData.Value(), &users); err != nil {
      return ctrl.Result{}, err
    }

    delete(users, instance.Name)
    usersJson, err := json.Marshal(users)
    if err != nil {
      return ctrl.Result{}, err
    }

    _, err = kv.Put("credentials", []byte(usersJson))
    if err != nil {
      return ctrl.Result{}, err
    }

    controllerutil.RemoveFinalizer(instance, finalizer)
  }

  password, err := GenerateRandomPassword(32)
  if err != nil {
    return ctrl.Result{}, err
  }

  hashedPassword, err := HashPassword(password)
  if err != nil {
    return ctrl.Result{}, err
  }

  data := map[string]any{
    "account": instance.Spec.Account,
    "pass": hashedPassword,
    "permissions": map[string]any{
      "pub": map[string]any{
        "allow": instance.Spec.Permissions.Pub.Allow,
        "deny": instance.Spec.Permissions.Pub.Deny,
      },
      "sub": map[string]any{
        "allow": instance.Spec.Permissions.Sub.Allow,
        "deny": instance.Spec.Permissions.Sub.Deny,
      },
    },
  }

  rawData, err := kv.Get("credentials")
  var users map[string]any
  if err := json.Unmarshal(rawData.Value(), &users); err != nil {
    return ctrl.Result{}, err
  }

  users[instance.Name] = data
  jsonUsers, err := json.Marshal(users)
  if err != nil {
    return ctrl.Result{}, err
  }

  _, err = kv.Put("credentials", jsonUsers)
  if err != nil {
    return ctrl.Result{}, err
  }

	return ctrl.Result{}, nil
}

func HashPassword(password string) (string, error) {
  bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
  return string(bytes), err
}
func GenerateRandomPassword(length int) (string, error) {
  randomBytes := make([]byte, length)
  _, err := rand.Read(randomBytes)
  if err != nil {
    return "", err
  }
  return base64.StdEncoding.EncodeToString(randomBytes), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NATSUserAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1alpha1.NATSUserAccount{}).
		Complete(r)
}
